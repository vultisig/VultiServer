import React, { useState } from "react";
import { initWasm, TW } from "@trustwallet/wallet-core";

const TokenSend: React.FC = () => {
  const [vaultAddress, setVaultAddress] = useState<string>("");
  const [balance, setBalance] = useState<string>("");
  const [toAddress, setToAddress] = useState<string>("");
  const [amount, setAmount] = useState<string>("");

  const rpcEndpoint = "https://thornode.ninerealms.com";

  const getBalance = async () => {
    if (!vaultAddress) return;

    try {
      const response: any = await fetch(
        `${rpcEndpoint}/cosmos/bank/v1beta1/balances/${vaultAddress}`
      );
      const balances = response.data.balances;
      const runeBalance = balances.find((bal: any) => bal.denom === "rune");
      setBalance(runeBalance ? runeBalance.amount : "0");
    } catch (error) {
      console.error("Error fetching balance:", error);
    }
  };

  const sendTransaction = async () => {
    if (!vaultAddress || !toAddress || !amount) return;

    try {
      const walletCore = await initWasm();
      const accountInfo: any = await fetch(
        `${rpcEndpoint}/auth/accounts/${vaultAddress}`
      );
      const accountNumber = accountInfo.data.result.value.account_number;
      const sequence = accountInfo.data.result.value.sequence;

      //   const tx = TW.THORChainSwap.Proto.SwapInput.create({
      //     fromAddress: vaultAddress,
      //     toAddress: toAddress,
      //     fromAmount: amount,
      //     // fromAsset: 'rune',
      //     // memo: '',
      //   });
      //   const tx = create({
      //     accountNumber: accountNumber,
      //     chainId: "thorchain",
      //     fee: {
      //       amounts: [{ denom: "rune", amount: "2000" }],
      //       gas: "200000",
      //     },
      //     memo: "",
      //     msgs: [
      //       {
      //         type: "cosmos-sdk/MsgSend",
      //         value: {
      //           from_address: vaultAddress,
      //           to_address: toAddress,
      //           amount: [{ denom: "rune", amount: amount }],
      //         },
      //       },
      //     ],
      //     sequence: sequence,
      //   });

      //   const encodedTx = tx.encode();
      //   const unsignedTx = Buffer.from(encodedTx).toString("hex");

      // api call

      //   const response = await provider.sendTransaction(signedTx);
      //   setTransactionHash(response.hash);
    } catch (error) {
      console.error("Error sending transaction:", error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
      <div className="bg-white p-6 rounded-lg shadow-md w-96">
        <h1 className="text-xl font-bold mb-4">Thorchain Wallet Demo</h1>
        <div className="mb-4">
          <label className="block text-gray-700">Wallet Address</label>
          <input
            type="text"
            value={vaultAddress}
            onChange={(e) => setVaultAddress(e.target.value)}
            className="mt-1 p-2 w-full border border-gray-300 rounded-md"
          />
        </div>
        <button
          onClick={getBalance}
          className="w-full bg-blue-500 text-white py-2 rounded-md mb-4 hover:bg-blue-600"
        >
          Get Balance
        </button>
        {balance && (
          <div className="mb-4">
            <label className="block text-gray-700">Balance</label>
            <p className="mt-1 p-2 w-full border border-gray-300 rounded-md bg-gray-50">
              {balance} RUNE
            </p>
          </div>
        )}
        <div className="mb-4">
          <label className="block text-gray-700">To Address</label>
          <input
            type="text"
            value={toAddress}
            onChange={(e) => setToAddress(e.target.value)}
            className="mt-1 p-2 w-full border border-gray-300 rounded-md"
          />
        </div>
        <div className="mb-4">
          <label className="block text-gray-700">Amount (RUNE)</label>
          <input
            type="text"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="mt-1 p-2 w-full border border-gray-300 rounded-md"
          />
        </div>
        <button
          onClick={sendTransaction}
          className="w-full bg-green-500 text-white py-2 rounded-md hover:bg-green-600"
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default TokenSend;
