use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint::ProgramResult,
    program_error::ProgramError,
    msg,
    pubkey::Pubkey,
    program_pack::{Pack, IsInitialized},
    sysvar::{rent::Rent, Sysvar},
    program::{invoke, invoke_signed},
    // spl_token::account as TokenAccount
};

use crate::{instruction::EscrowInstruction, error::EscrowError, state::Escrow};

use spl_token::state::Account as TokenAccount;

pub struct Processor;
impl Processor {
    pub fn process(program_id: &Pubkey, accounts: &[AccountInfo], instruction_data: &[u8]) -> ProgramResult {
        let instruction = EscrowInstruction::unpack(instruction_data)?;

        match instruction {
            EscrowInstruction::InitEscrow { amount } => {
                msg!("Instruction: InitEscrow");
                Self::process_init_escrow(accounts, amount, program_id)
            },
            EscrowInstruction::Exchange { amount } => {
                msg!("Instruction: Exchange");
                Self::process_exchange(accounts, amount, program_id)
            }
        }
    }


    /// IMPORTANT SECURITY QUOTE:
    // There is a bug in this program. It's nothing critical. I've left it in because it does showcase the subtleties of programming on Solana. Can you find it?

    // It's inside process_init_escrow

    // There's an additional check missing that has to do with the possible states of an account

    // We check that Alice's Y token account is owned by the token program but what we don't check is that the given account is actually a token account. It could also be a token mint account. This is not a critical bug because Bob's tx will simply fail (when the program tries to transfer his Y tokens to a mint account) but for the same reasons as the check for correct program ownership we should add this check in Alice's ix.
    fn process_init_escrow(accounts: &[AccountInfo], amount:u64, program_id: &Pubkey) -> ProgramResult {
        let account_info_iter = &mut accounts.iter();
        let initializer = next_account_info(account_info_iter)?;

        if !initializer.is_signer {
            return Err(ProgramError:: MissingRequiredSignature);
        }

        let temp_token_account = next_account_info(account_info_iter)?;
        // implicit account writable and ownership check

        let token_to_receive_account = next_account_info(account_info_iter)?;
        if *token_to_receive_account.owner != spl_token::id() {
            return Err(ProgramError::IncorrectProgramId);
        }

        let escrow_account = next_account_info(account_info_iter)?;
        let rent = &Rent::from_account_info(next_account_info(account_info_iter)?)?;

        if !rent.is_exempt(escrow_account.lamports(), escrow_account.data_len()) {
            return Err(EscrowError::NotRentExempt.into());
        }

        let mut escrow_info = Escrow::unpack_unchecked(&escrow_account.try_borrow_data()?)?;
        if escrow_info.is_initialized() {
            return Err(ProgramError::AccountAlreadyInitialized);
        }

        escrow_info.is_initialized = true;
        escrow_info.initializer_pubkey = *initializer.key;
        escrow_info.temp_token_account_pubkey = *temp_token_account.key;
        escrow_info.initializer_token_to_receive_account_pubkey = *token_to_receive_account.key;
        escrow_info.expected_amount = amount;

        Escrow::pack(escrow_info, &mut escrow_account.try_borrow_mut_data()?)?;
        let (pda, _bump_seed) = Pubkey::find_program_address(&[b"escrow"], program_id);

        let token_program = next_account_info(account_info_iter)?;
        let owner_change_ix = spl_token::instruction::set_authority(
            token_program.key,
            temp_token_account.key,
            Some(&pda),
            spl_token::instruction::AuthorityType::AccountOwner,
            initializer.key,
            &[&initializer.key],
        )?;

        msg!("Callign the token program to transfer token account ownership...");
        //automatically checks if the token program is the right and a rogue one
        invoke(
            &owner_change_ix,
            &[
                temp_token_account.clone(),
                initializer.clone(),
                token_program.clone(),
            ],
        )?;

        Ok(())
    }

    //IMPORTANT QUOTE:
    // add a Cancel endpoint to the program. Currently, Alice's tokens are stuck in limbo and she will not be able to recover them if Bob decides not to take the trade. Add an endpoint that allows Alice to cancel the ongoing escrow, transferring the X tokens back to her and closing the two created accounts. 🚨 If you implement cancel, you also need to add another check to prevent a frontrunning attack. Preventing it requires that Bob also sends the amount of Y tokens that he expects to send Alice (expected_y_amount) in addition to the amount of X tokens he expects from her. The check belongs in process_exchange and verifies that escrow_info.expected_amount == expected_y_amount. This prevents the following attack: Once Bob sends his Transaction and Alice sees it, Alice can cancel the escrow, reinitialise it at the same address but with a higher expected amount, thereby receiving more Y tokens than Bob expected to give her. Alternatively (or additionally), Bob could use a temporary token account himself so that in the case Alice frontruns him, his tx will fail because there's not enough Y tokens in his temporary token account. In addition, you also need to have Bob send his expected_x_amount (see the frontrunning section in #instruction-rs-part-3-understanding-what-bob-s-transaction-should-do (opens new window)).

    fn process_exchange(
        accounts: &[AccountInfo],
        amount_expected_by_taker: u64,
        program_id: &Pubkey,
    ) -> ProgramResult {
        let account_info_iter = &mut accounts.iter();
        let taker = next_account_info(account_info_iter)?;

        if !taker.is_signer {
            return Err(ProgramError::MissingRequiredSignature);
        }

        let takers_sending_token_account = next_account_info(account_info_iter)?;

        let takers_token_to_receive_account = next_account_info(account_info_iter)?;

        let pdas_temp_token_account = next_account_info(account_info_iter)?;
        let pdas_temp_token_account_info = TokenAccount::unpack(&pdas_temp_token_account.try_borrow_data()?)?;
        let (pda, bump_seed) = Pubkey::find_program_address(&[b"escrow"], program_id);

        if amount_expected_by_taker != pdas_temp_token_account_info.amount
        {
            return Err(EscrowError::ExpectedAmountMismatch.into());
        }  

        let initializers_main_account = next_account_info(account_info_iter)?;
        let initializers_token_to_receive_account = next_account_info(account_info_iter)?;
        let escrow_account = next_account_info(account_info_iter)?;

        let escrow_info = Escrow::unpack(&escrow_account.try_borrow_data()?)?;

        if escrow_info.temp_token_account_pubkey != *pdas_temp_token_account.key {
            return Err(ProgramError::InvalidAccountData);
        }

        if escrow_info.initializer_pubkey != *initializers_main_account.key {
            return Err(ProgramError::InvalidAccountData);
        }

        if escrow_info.initializer_token_to_receive_account_pubkey != *initializers_token_to_receive_account.key {
            return Err(ProgramError::InvalidAccountData);
        }

        let token_program = next_account_info(account_info_iter)?;

        let transfer_to_initializer_ix = spl_token::instruction::transfer(
           token_program.key,
           takers_sending_token_account.key,
           initializers_token_to_receive_account.key,
           taker.key,
           &[&taker.key],
           escrow_info.expected_amount, 
        )?;

        msg!("Calling the token program to transfer tokens to the escrow's initializer...");
        invoke(
            &transfer_to_initializer_ix,
            &[
                takers_sending_token_account.clone(),
                initializers_token_to_receive_account.clone(),
                taker.clone(),
                token_program.clone(),
            ],
        )?;

        let pda_account = next_account_info(account_info_iter)?;

        let transfer_to_taker_ix = spl_token::instruction::transfer(
            token_program.key,
            pdas_temp_token_account.key,
            takers_token_to_receive_account.key,
            &pda,
            &[&pda],
            pdas_temp_token_account_info.amount,
        )?;
        msg!("Calling the token program to transfer tokens to the taker...");
        invoke_signed(
            &transfer_to_taker_ix,
            &[
                pdas_temp_token_account.clone(),
                takers_token_to_receive_account.clone(),
                pda_account.clone(),
                token_program.clone(),
            ],
            &[&[&b"escrow"[..], &[bump_seed]]],
        )?;

        let close_pdas_temp_acc_ix = spl_token::instruction::close_account(
            token_program.key,
            pdas_temp_token_account.key,
            initializers_main_account.key,
            &pda,
            &[&pda]
        )?;
        msg!("Calling the token program to close pda's temp account...");
        invoke_signed(
            &close_pdas_temp_acc_ix,
            &[
                pdas_temp_token_account.clone(),
                initializers_main_account.clone(),
                pda_account.clone(),
                token_program.clone(),
            ],
            &[&[&b"escrow"[..], &[bump_seed]]],
        )?;

        msg!("Closing the escrow account...");
        **initializers_main_account.lamports.borrow_mut() = initializers_main_account.lamports().checked_add(escrow_account.lamports()).ok_or(EscrowError::AmountOverflow)?;
        **escrow_account.lamports.borrow_mut() = 0;
        //IMPORTANT QUOTE: Depending on your program, forgetting to clear the data field can have dangerous consequences.It is because this instruction is not necessarily the final instruction in the transaction. Thus, a subsequent transaction may read or even revive the data completely by making the account rent-exempt again.
        *escrow_account.try_borrow_mut_data()? = &mut [];

        Ok(())
    }
}