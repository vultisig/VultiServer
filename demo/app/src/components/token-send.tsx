import React, { useState } from "react";
import { initWasm, TW, WalletCore } from "@trustwallet/wallet-core";

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

  function getPreSignedInputData(
    walletCore: WalletCore,
    keysignPayload: KeysignPayload
  ): Uint8Array {
    const { AnyAddress } = walletCore;

    const fromAddr = AnyAddress.createWithString(
      keysignPayload.coin.address,
      walletCore.CoinType.thorchain
    );
    if (!fromAddr) {
      console.error(`${keysignPayload.coin.address} is invalid`);
      return Uint8Array.of();
    }

    if (!keysignPayload.chainSpecific) {
      console.error("fail to get account number, sequence, or fee");
      return Uint8Array.of();
    }

    const { accountNumber, sequence } = keysignPayload.chainSpecific;
    const pubKeyData = Buffer.from(keysignPayload.coin.hexPublicKey, "hex");
    if (!pubKeyData) {
      console.error("invalid hex public key");
      return Uint8Array.of();
    }

    const toAddress = AnyAddress.createWithString(
      keysignPayload.toAddress,
      walletCore.CoinType.thorchain
    );
    if (!toAddress) {
      console.error(`${keysignPayload.toAddress} is invalid`);
      return Uint8Array.of();
    }

    const message = [
      TW.Cosmos.Proto.Message.create({
        thorchainSendMessage: TW.Cosmos.Proto.Message.THORChainSend.create({
          fromAddress: fromAddr.data(),
          amounts: [
            TW.Cosmos.Proto.Amount.create({
              denom: "rune",
              amount: keysignPayload.toAmount.toString(),
            }),
          ],
          toAddress: toAddress.data(),
        }),
      }),
    ];

    let chainId = walletCore.CoinTypeExt.chainId(walletCore.CoinType.thorchain);
    // if (chainId !== walletCore.Blockchain.thorchain.value.toString()) {
    //   chainId = walletCore.Blockchain.thorchain.value.toString();
    // }

    const input = TW.Cosmos.Proto.SigningInput.create({
      publicKey: pubKeyData,
      signingMode: TW.Cosmos.Proto.SigningMode.Protobuf,
      chainId: chainId,
      accountNumber: accountNumber,
      sequence: sequence,
      mode: TW.Cosmos.Proto.BroadcastMode.SYNC,
      memo: keysignPayload.memo,
      messages: message,
      fee: { gas: 20000000 },
    });

    return TW.Cosmos.Proto.SigningInput.encode(input).finish();
  }

  function getSignedTransaction(
    walletCore: WalletCore,
    vaultHexPubKey: string,
    vaultHexChainCode: string,
    keysignPayload: KeysignPayload,
    signatures: { [key: string]: TssKeysignResponse }
  ) {
    const inputData = getPreSignedInputData(walletCore, keysignPayload);
    const thorPublicKey = getDerivedPubKey(
      vaultHexPubKey,
      vaultHexChainCode,
      walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain)
    ); // need to get it from tss service
    const pubkeyData = Buffer.from(thorPublicKey, "hex");
    const publicKey = walletCore.PublicKey.createWithData(
      pubkeyData,
      walletCore.PublicKeyType.secp256k1
    );

    if (!publicKey) {
      console.error(`public key ${thorPublicKey} is invalid`);
      return;
    }

    try {
      const hashes = walletCore.TransactionCompiler.preImageHashes(
        walletCore.CoinType.thorchain,
        inputData
      );
      const preSigningOutput =
        TW.TxCompiler.Proto.PreSigningOutput.decode(hashes);
      const allSignatures = new walletCore.DataVector();
      const publicKeys = new walletCore.DataVector();
      const signature = signatures[preSigningOutput.dataHash.toString()];

      if (!publicKey.verify(signature, preSigningOutput.dataHash)) {
        console.error("fail to verify signature");
        return;
      }

      allSignatures.add(signature);
      publicKeys.add(pubkeyData);

      const compileWithSignature =
        walletCore.TransactionCompiler.compileWithSignatures(
          walletCore.CoinType.thorchain,
          inputData,
          allSignatures,
          publicKeys
        );
      const output = TW.Cosmos.Proto.SigningOutput.decode(compileWithSignature);
      return output.serialized; // raw transaction
    } catch (error: any) {
      console.error(
        `fail to get signed ethereum transaction, error: ${error.message}`
      );
      return;
    }
  }

  async function broadcastSignedTransaction(signedTx: string): Promise<any> {
    try {
      const response = await fetch(`${rpcEndpoint}/broadcast`, {
        method: "POST",
        body: signedTx,
      });

      if (response.status === 200) {
        const data = await response.json();
        console.log("Transaction broadcasted successfully:", data);
        return data;
      } else {
        console.error(
          "Failed to broadcast transaction:",
          response.status,
          response.statusText
        );
        return;
      }
    } catch (error) {
      console.error("Error broadcasting transaction:", error);
      return;
    }
  }

  const sendTransaction = async () => {
    if (!vaultAddress || !toAddress || !amount) return;

    try {
      const walletCore = await initWasm();
      const accountInfo: any = await fetch(
        `${rpcEndpoint}/auth/accounts/${vaultAddress}`
      );
      const accountNumber = accountInfo.data.result.value.account_number;
      const sequence = accountInfo.data.result.value.sequence;

      // build keysign request
      // api call to get keysignPayload, chain code
      // api call to get signatures after task completed

      //   const signedTx = getSignedTransaction(walletCore, vaultAddress, vaultHexChainCode, keysignPayload, signatures)
      //   if (signedTx) {
      //     broadcastSignedTransaction(signedTx)
      //   }
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
