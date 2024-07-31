import React, { useState } from "react";
import { initWasm, TW, WalletCore } from "@trustwallet/wallet-core";
import { v4 as uuidv4 } from "uuid";
import { endPoints } from "../api/endpoints";
import {
  Coin,
  KeysignMessage,
  KeysignPayload,
  KeysignPayloadType,
  KeysignResponse,
  THORChainSpecific,
} from "../utils/types";
import { createHash } from "crypto";
import { QRCode } from "react-qrcode-logo";
import { Buffer } from "buffer";
import {
  byteArrayToHexString,
  generateRandomHex,
  getSignatureWithRecoveryID,
  lzmaCompressData,
} from "../utils/utils";

const TokenSend: React.FC = () => {
  const [vaultPublicKeyEcdsa, setVaultPublicKeyEcdsa] = useState<string>("");
  const [balance, setBalance] = useState<string>("");
  const [toAddress, setToAddress] = useState<string>("");
  const [amount, setAmount] = useState<string>("");
  const [passwd, setPasswd] = useState<string>("");
  const [qrString, setQrString] = useState<string>("");

  const rpcEndpoint = "https://thornode.ninerealms.com";

  const getDerivedPublicKey = async (
    publicKey: string,
    hexChainCode: string,
    derivePath: string,
    isEdDSA: boolean
  ) => {
    if (!publicKey) return;

    const queryParams = new URLSearchParams({
      publicKey,
      hexChainCode,
      derivePath,
      isEdDSA: isEdDSA ? "true" : "",
    }).toString();
    try {
      const response = await fetch(
        `${endPoints.getDerivedPublicKey}?${queryParams}`
      );
      const derivedPublicKey = await response.json();

      return derivedPublicKey;
    } catch (error) {
      console.error("Error getThorchainPublicKey:", error);
    }
  };

  const getBalance = async () => {
    if (!vaultPublicKeyEcdsa) return;

    try {
      let response: any = await fetch(
        `${endPoints.getVault}/${vaultPublicKeyEcdsa}`,
        {
          headers: {
            "x-password": passwd,
          },
        }
      );
      const vaultInfo = await response.json();
      const vaultHexChainCode = vaultInfo.hex_chain_code;

      const walletCore = await initWasm();
      const thorPublicKeyStr = await getDerivedPublicKey(
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
        false
      );
      const pubkeyData = Buffer.from(thorPublicKeyStr, "hex");
      const thorPublicKey = walletCore.PublicKey.createWithData(
        pubkeyData,
        walletCore.PublicKeyType.secp256k1
      );
      const thorAddress = walletCore.CoinTypeExt.deriveAddressFromPublicKey(
        walletCore.CoinType.thorchain,
        thorPublicKey
      );
      response = await fetch(
        `${rpcEndpoint}/cosmos/bank/v1beta1/balances/${thorAddress}`
      );
      const data = await response.json();
      const runeBalance = data.balances.find(
        (bal: any) => bal.denom === "rune"
      );
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
    const { account_number, sequence } = keysignPayload.thorchain_specific;
    const pubKeyData = Buffer.from(keysignPayload.coin.hex_public_key, "hex");
    const pubKey = walletCore.PublicKey.createWithData(
      pubKeyData,
      walletCore.PublicKeyType.secp256k1
    );
    const toAddress = AnyAddress.createWithString(
      keysignPayload.to_address,
      walletCore.CoinType.thorchain
    );
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
      fee: TW.Cosmos.Proto.Fee.create({
        gas: 2000000,
      }),
    });

    return TW.Cosmos.Proto.SigningInput.encode(input).finish();
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
        signatures[byteArrayToHexString(preSigningOutput.dataHash)]
      );
      console.log(
        222,
        preSigningOutput,
        preSigningOutput.dataHash,
        byteArrayToHexString(preSigningOutput.dataHash)
      );
      console.log(333, "signature", signature);
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

      {
        const buffer = Buffer.from(output.serialized, "base64");
        const hash = createHash("sha256").update(buffer).digest("hex");
        console.log("txHash:", hash);
      }

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
      let response: any = await fetch(
        `${endPoints.getVault}/${vaultPublicKeyEcdsa}`,
        {
          headers: {
            "x-password": passwd,
          },
        }
      );
      const vaultInfo = await response.json();
      const vaultLocalPartyId = vaultInfo.local_party_id;
      const vaultHexChainCode = vaultInfo.hex_chain_code;
      const thorPublicKeyStr = await getDerivedPublicKey(
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        walletCore.CoinTypeExt.derivationPath(walletCore.CoinType.thorchain),
        false
      );
      const pubkeyData = Buffer.from(thorPublicKeyStr, "hex");
      const thorPublicKey = walletCore.PublicKey.createWithData(
        pubkeyData,
        walletCore.PublicKeyType.secp256k1
      );
      const thorAddress = walletCore.CoinTypeExt.deriveAddressFromPublicKey(
        walletCore.CoinType.thorchain,
        thorPublicKey
      );

      response = await fetch(`${rpcEndpoint}/auth/accounts/${thorAddress}`);
      const accountInfo = await response.json();
      const accountNumber = accountInfo.result.value.account_number;
      const sequence = accountInfo.result.value.sequence;
      const thorchainspecific = new THORChainSpecific({
        account_number: accountNumber,
        sequence: sequence,
      });
      const coin = new Coin({
        chain: "THOR",
        ticker: "RUNE",
        is_native_token: true,
        hex_public_key: vaultPublicKeyEcdsa,
        address: thorAddress,
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
      const message = byteArrayToHexString(preSigningOutput.signature);

      const sessionId = uuidv4();
      const hexEncryptionKey = generateRandomHex(32);

      const keysignMsg = new KeysignMessage({
        session_id: sessionId,
        service_name: "VultisignerApp",
        keysign_payload: payload,
        encryption_key_hex: hexEncryptionKey,
        use_vultisig_relay: true,
      });
      const compressedData = await lzmaCompressData(keysignMsg.serialize());

      setQrString(
        `vultisig://vultisig.com?type=SignTransaction&vault=${vaultPublicKeyEcdsa}&jsonData=${Buffer.from(
          compressedData
        ).toString("base64")}`
      );

      // build util function like createVault
      response = await fetch(endPoints.sign, {
        method: "POST",
        headers: {
          "x-password": passwd,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          public_key: vaultPublicKeyEcdsa,
          messages: [message],
          session: sessionId,
          hex_encryption_key: hexEncryptionKey,
          derive_path: walletCore.CoinTypeExt.derivationPath(
            walletCore.CoinType.thorchain
          ),
          is_ecdsa: true,
          vault_password: "",
        }),
      });
      const taskId = await response.json();
      setTimeout(() => {
        signTransaction(taskId, walletCore, vaultHexChainCode, payload);
      }, 20000);
    } catch (error) {
      console.error("Error sending transaction:", error);
    }
  };

  const signTransaction = async (
    taskId: string,
    walletCore: WalletCore,
    vaultHexChainCode: string,
    payload: KeysignPayloadType
  ) => {
    try {
      const resp = await fetch(`${endPoints.getSignResult}/${taskId}`);
      if (resp.status !== 200) {
        console.error("Invalid Task Id");
        setTimeout(() => {
          signTransaction(taskId, walletCore, vaultHexChainCode, payload);
        }, 5000);
        return;
      }
      const signatures = await resp.json();
      console.log(111, "signatures", signatures);
      if (
        signatures.message &&
        signatures.message === "Task is still in progress"
      ) {
        setTimeout(() => {
          signTransaction(taskId, walletCore, vaultHexChainCode, payload);
        }, 5000);
        return;
      }

      const signedTx = await getSignedTransaction(
        walletCore,
        vaultPublicKeyEcdsa,
        vaultHexChainCode,
        payload,
        signatures
      );
      console.log(444, "signedTx", signedTx);
      if (signedTx) {
        broadcastSignedTransaction(signedTx);
      }
    } catch (error) {
      console.error("Error signing transaction:", error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
      <div className="bg-white p-6 rounded-lg shadow-md w-96 text-black">
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
      {qrString && (
        <div className="bg-white p-4 rounded-lg mx-8">
          <QRCode
            value={qrString}
            size={250}
            bgColor={"#ffffff"}
            fgColor={"#0B51C6"}
            ecLevel="L"
            logoImage="/img/logo.svg"
            logoWidth={70}
            logoHeight={70}
            qrStyle="dots"
            eyeRadius={[
              {
                outer: [10, 10, 0, 10],
                inner: [0, 10, 10, 10],
              },
              [10, 10, 10, 0],
              [10, 0, 10, 10],
            ]}
          />
        </div>
      )}
    </div>
  );
};

export default TokenSend;
