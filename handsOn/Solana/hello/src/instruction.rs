use std::convert::TryInto;
use solana_program::program_error::ProgramError;

use crate::error::EscrowError::InvalidInstruction;

pub enum EscrowInstruction {
    
    /// Starts the trade by creating and populating an escrow account and transferring ownership of the given temp token account to the PDA
    ///
    ///
    /// Accounts expected:
    ///
    /// 0. `[signer]` The account of the person initializing the escrow
    /// 1. `[writable]` Temporary token account that should be created prior to this instruction and owned by the initializer
    /// 2. `[]` The initializer's token account for the token they will receive should the trade go through
    /// 3. `[writable]` The escrow account, it will hold all necessary info about the trade.
    /// 4. `[]` The rent sysvar
    /// 5. `[]` The token program
    InitEscrow {
        /// The amount party A expects to receive of token Y
        amount: u64
    },

    /// Accepts a trade
    ///
    ///
    /// Accounts expected:
    ///
    /// 0. `[signer]` The account of the person taking the trade
    /// 1. `[writable]` The taker's token account for the token they send 
    /// 2. `[writable]` The taker's token account for the token they will receive should the trade go through
    /// 3. `[writable]` The PDA's temp token account to get tokens from and eventually close
    /// 4. `[writable]` The initializer's main account to send their rent fees to
    /// 5. `[writable]` The initializer's token account that will receive tokens
    /// 6. `[writable]` The escrow account holding the escrow info
    /// 7. `[]` The token program
    /// 8. `[]` The PDA account
    Exchange {
        /// the amount the taker expects to be paid in the other token, as a u64 because that's the max possible supply of a token
        amount: u64,
    }

    // Qouting the post: IMPORTANT FOR SECURITY ANALYSIS
    // Importantly, the ix also expects an amount. Why is this necessary? After all, Bob could simply look up the escrow information account in the explorer and check that in its state, it has a reference to a temporary token account with the amount of X tokens he is comfortable receiving for his Y tokens. Why should the amount he expects still be included in his ix? Try to figure it out yourself!

    // it's only a problem if we also implement a cancel endpoint that allows Alice to cancel an escrow she created

    // frontrunning

    // Leaving out the amount makes possible a specific attack on the program. Note that this attack would only work if we added a cancel ix to the program that completely closed the escrow information account, allowing it to be reinitialized as a new escrow account. If this is implemented, Bob's ix can be frontrun. There already exist good explanations of frontrunning and frontrunning in crypto so I will not explain it in general here and instead focus on our program. Let's assume Alice opens a new escrow, offering to exchange her 10 SOL for someone's 2000USDC (Note that the token program has special functionality to convert native SOL into a SOL token, this makes it easy for other programs - like ours - to handle SOL without having to build extra functionality to handle native SOL). Bob sees the escrow on-chain and likes the terms of the escrow. He decides to send a process_exchange instruction. Importantly, in this version of the escrow, there is no check in the program that verifies that the amount Bob expects equals the amount in the escrow information account because Bob does not send the amount he expects in the first place. This means that by sending his ix, he agrees to whatever is written in the escrow information account. He thinks this is fine because before accepting he looked at the data field in the explorer and it contained a temp X account with the amount he expected. However, Alice now has a way to take his tokens without giving up hers. She could, for example, do this by colluding with the current slot leader, who is responsible for ordering the tx in their assigned slot. After Bob sent his tx to the leader, Alice's sends a tx that cancels the escrow and immediately creates a new one at the same address which now holds a reference to an empty temporary X token account. The leader puts this tx before Bob's. As a result, by the time Bob's tx is processed, the escrow Bob's tx references will reference the empty temporary X token account, so Bob will not receive anything but Alice will still receive Bob's Y tokens. Adding the expected amount in Bob's instruction will prevent this issue. Bob should, of course, verify that the program uses this information to check that tempXtokenAccount.amount == expectedAmount and does not just ignore it.
}

impl EscrowInstruction {
    pub fn unpack(input: &[u8]) -> Result<Self, ProgramError> {
        let (tag, rest) = input.split_first().ok_or(InvalidInstruction)?;

        Ok(match tag {
            0 => Self::InitEscrow {
                amount: Self::unpack_amount(rest)?,
            },
            1 => Self::Exchange {
                amount: Self::unpack_amount(rest)?
            },
        _ => return Err(InvalidInstruction.into()),
        })
    }

    fn unpack_amount(input: &[u8]) -> Result<u64, ProgramError> {
        let amount = input.get(..8).and_then(|slice| slice.try_into().ok()).map(u64::from_le_bytes).ok_or(InvalidInstruction)?;
        Ok(amount)
    }
}