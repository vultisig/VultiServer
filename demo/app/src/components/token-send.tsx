import React, { useState } from "react";
import { initWasm, TW, WalletCore } from "@trustwallet/wallet-core";
import { v4 as uuidv4 } from "uuid";
import { endPoints } from "../api/endpoints";
import {
  Coin,
  KeysignPayload,
  KeysignPayloadType,
  KeysignResponse,
  THORChainSpecific,
} from "../utils/types";
import { randomBytes } from "crypto";

const TokenSend: React.FC = () => {
  const [vaultPublicKeyEcdsa, setVaultPublicKeyEcdsa] = useState<string>("");
  const [balance, setBalance] = useState<string>("");
  const [toAddress, setToAddress] = useState<string>("");
  const [amount, setAmount] = useState<string>("");
  const [passwd, setPasswd] = useState<string>("");

  const rpcEndpoint = "https://thornode.ninerealms.com";

  const getDerivedPublicKey = async (
    publicKey: string,
    hexChainCode: string,
    derivePath: string,
    isEdDSA: boolean
  ) => {
    if (!publicKey) return;

    try {
      const response = await fetch(endPoints.getDerivedPublicKey, {
        body: JSON.stringify({
          public_key: publicKey,
          hex_chain_code: hexChainCode,
          derive_path: derivePath,
          is_eddsa: isEdDSA,
        }),
      });
      const data = await response.json();
      return data.derivedPublicKey;
    } catch (error) {
      console.error("Error getThorchainPublicKey:", error);
    }
  };

  const getBalance = async () => {
    if (!vaultPublicKeyEcdsa) return;

    try {
      const vaultInfo: any = await fetch(
        `${endPoints.getVault}/${vaultPublicKeyEcdsa}`,
        {
          headers: {
            "x-password": passwd,
          },
        }
      );
      const vaultHexChainCode = vaultInfo.data.result.value.hex_chain_code;

      const walletCore = await initWasm();
      const thorPublicKey = await getDerivedPublicKey(
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
        false
      );
      // const pubkeyData = Buffer.from(thorPublicKey, "hex");
      // const thorAddress = walletCore.PublicKey.createWithData(
      //   pubkeyData,
      //   walletCore.PublicKeyType.secp256k1
      // );
      const response: any = await fetch(
        `${rpcEndpoint}/cosmos/bank/v1beta1/balances/${thorPublicKey}`
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
    keysignPayload: KeysignPayloadType
  ): Uint8Array {
    const { AnyAddress } = walletCore;

    const fromAddr = AnyAddress.createWithString(
      keysignPayload.coin.address,
      walletCore.CoinType.thorchain
    );
    // if (!fromAddr) {
    //   console.error(`${keysignPayload.coin.address} is invalid`);
    //   return Uint8Array.of();
    // }

    if (!keysignPayload.thorchain_specific) {
      console.error("fail to get account number, sequence, or fee");
      return Uint8Array.of();
    }

    const { account_number, sequence } = keysignPayload.thorchain_specific;
    const pubKeyData = Buffer.from(keysignPayload.coin.hex_public_key, "hex");
    const pubKey = walletCore.PublicKey.createWithData(
      pubKeyData,
      walletCore.PublicKeyType.secp256k1
    );
    // if (!pubKeyData) {
    //   console.error("invalid hex public key");
    //   return Uint8Array.of();
    // }

    const toAddress = AnyAddress.createWithString(
      keysignPayload.to_address,
      walletCore.CoinType.thorchain
    );
    // if (!toAddress) {
    //   console.error(`${keysignPayload.to_address} is invalid`);
    //   return Uint8Array.of();
    // }

    const messages = [
      TW.Cosmos.Proto.Message.create({
        thorchainSendMessage: TW.Cosmos.Proto.Message.THORChainSend.create({
          fromAddress: fromAddr.data(),
          amounts: [
            TW.Cosmos.Proto.Amount.create({
              denom: keysignPayload.coin.ticker.toLowerCase(),
              amount: keysignPayload.to_amount.toString(),
            }),
          ],
          toAddress: toAddress.data(),
        }),
      }),
    ];

    const input = TW.Cosmos.Proto.SigningInput.create({
      publicKey: pubKey.data(),
      signingMode: TW.Cosmos.Proto.SigningMode.Protobuf,
      chainId: walletCore.CoinTypeExt.chainId(walletCore.CoinType.thorchain),
      accountNumber: account_number,
      sequence: sequence,
      mode: TW.Cosmos.Proto.BroadcastMode.SYNC,
      memo: keysignPayload.memo,
      messages: messages,
      fee: { gas: 20000000 },
    });

    return TW.Cosmos.Proto.SigningInput.encode(input).finish();
  }

  function hexStringToByteArray(hex: string): Uint8Array {
    if (hex.length % 2 !== 0) {
      throw new Error("Invalid hex string");
    }
    const byteArray = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      byteArray[i / 2] = parseInt(hex.substr(i, 2), 16);
    }
    return byteArray;
  }

  function getSignatureWithRecoveryID(signature: KeysignResponse): Uint8Array {
    const rBytes = hexStringToByteArray(signature.r);
    const sBytes = hexStringToByteArray(signature.s);
    const recoveryIdBytes = hexStringToByteArray(signature.recovery_id);
    const signatureWithRecoveryID = new Uint8Array(
      rBytes.length + sBytes.length + recoveryIdBytes.length
    );
    signatureWithRecoveryID.set(rBytes);
    signatureWithRecoveryID.set(sBytes, rBytes.length);
    signatureWithRecoveryID.set(recoveryIdBytes, rBytes.length + sBytes.length);

    return signatureWithRecoveryID;
  }

  const getSignedTransaction = async (
    walletCore: WalletCore,
    vaultHexPubKey: string,
    vaultHexChainCode: string,
    keysignPayload: KeysignPayloadType,
    signatures: { [key: string]: KeysignResponse }
  ) => {
    const inputData = getPreSignedInputData(walletCore, keysignPayload);
    const thorPublicKey = await getDerivedPublicKey(
      vaultHexPubKey,
      vaultHexChainCode,
      walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
      false
    );
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
      const allPublicKeys = new walletCore.DataVector();
      const signature = getSignatureWithRecoveryID(
        signatures[preSigningOutput.dataHash.toString()]
      ); // to hex string?

      if (!publicKey.verify(signature, preSigningOutput.dataHash)) {
        console.error("fail to verify signature");
        return;
      }

      allSignatures.add(signature);
      allPublicKeys.add(publicKey.data());

      const compileWithSignature =
        walletCore.TransactionCompiler.compileWithSignatures(
          walletCore.CoinType.thorchain,
          inputData,
          allSignatures,
          allPublicKeys
        );
      const output = TW.Cosmos.Proto.SigningOutput.decode(compileWithSignature);
      return output.serialized; // raw transaction
    } catch (error: any) {
      console.error(
        `fail to get signed ethereum transaction, error: ${error.message}`
      );
      return;
    }
  };

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
    if (!vaultPublicKeyEcdsa || !toAddress || !amount) return;

    try {
      const walletCore = await initWasm();
      const vaultInfo: any = await fetch(
        `${endPoints.getVault}/${vaultPublicKeyEcdsa}`,
        {
          headers: {
            "x-password": passwd,
          },
        }
      );
      const vaultLocalPartyId = vaultInfo.data.result.value.local_party_id;
      const vaultHexChainCode = vaultInfo.data.result.value.hex_chain_code;

      const thorPublicKey = await getDerivedPublicKey(
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
        false
      );
      // const pubkeyData = Buffer.from(thorPublicKey, "hex");
      // const thorAddress = walletCore.PublicKey.createWithData(
      //   pubkeyData,
      //   walletCore.PublicKeyType.secp256k1
      // );

      const accountInfo: any = await fetch(
        `${rpcEndpoint}/auth/accounts/${thorPublicKey}`
      );
      const accountNumber = accountInfo.data.result.value.account_number;
      const sequence = accountInfo.data.result.value.sequence;

      const thorchainspecific = new THORChainSpecific({
        account_number: accountNumber,
        sequence: sequence,
      });
      const coin = new Coin({
        chain: "THOR",
        ticker: "RUNE",
        is_native_token: true,
        hex_public_key: vaultPublicKeyEcdsa,
      });
      const payload = new KeysignPayload({
        coin: coin,
        to_address: toAddress,
        to_amount: amount,
        thorchain_specific: thorchainspecific,
        vault_public_key_ecdsa: vaultPublicKeyEcdsa,
        vault_local_party_id: vaultLocalPartyId,
      });
      const txInputData = getPreSignedInputData(walletCore, payload);
      const preImageHashes = walletCore.TransactionCompiler.preImageHashes(
        walletCore.CoinType.thorchain,
        txInputData
      );
      const preSigningOutput =
        TW.Cosmos.Proto.SigningOutput.decode(preImageHashes);
      const message = Array.from(preSigningOutput.signature)
        .map((byte) => byte.toString(16).padStart(2, "0"))
        .join("");

      await fetch(endPoints.sign, {
        method: "POST",
        headers: {
          "x-password": passwd,
        },
        body: JSON.stringify({
          public_key: vaultPublicKeyEcdsa,
          messages: [message],
          session: uuidv4(),
          hex_encryption_key: randomBytes(32).toString("hex"),
          derive_path: walletCore.CoinTypeExt.derivationPath(
            walletCore.CoinType.thorchain
          ),
          is_ecdsa: true,
        }),
      });

      // api call to get signatures after task completed
      const signatures = {};

      const signedTx = await getSignedTransaction(
        walletCore,
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        payload,
        signatures
      );
      if (signedTx) {
        broadcastSignedTransaction(signedTx);
      }
    } catch (error) {
      console.error("Error sending transaction:", error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
      <div className="bg-white p-6 rounded-lg shadow-md w-96">
        <h1 className="text-xl font-bold mb-4">Vault Token Send Demo</h1>
        <div className="mb-4">
          <label className="block text-gray-700">Vault PublicKey Ecdsa</label>
          <input
            type="text"
            value={vaultPublicKeyEcdsa}
            onChange={(e) => setVaultPublicKeyEcdsa(e.target.value)}
            className="mt-1 p-2 w-full border border-gray-300 rounded-md mb-4"
          />
          <input
            type="password"
            value={passwd}
            onChange={(e) => setPasswd(e.target.value)}
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
