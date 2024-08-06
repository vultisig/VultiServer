import { useState } from "react";
import { initWasm } from "@trustwallet/wallet-core";
import { Buffer } from "buffer";
import { getBalances } from "../../api/thorchain";
import { getDerivedPublicKey } from "../../api/utils/utils";
import { getVault } from "../../api/vault/vault";

interface StepOneProps {
  sendTransaction: (
    vaultPublicKeyEcdsa: string,
    vaultLocalPartyId: string,
    vaultHexChainCode: string,
    fromPublicKey: string,
    fromAddress: string,
    toAddress: string,
    amount: string,
    passwd: string
  ) => void;
}

const StepOne = ({ sendTransaction }: StepOneProps) => {
  const [vaultPublicKeyEcdsa, setVaultPublicKeyEcdsa] = useState<string>(
    "0282ee89003fbf985aaebff03d5c9b885a3d508c6fbd2974801e018a11bb79c01f"
  );
  const [balance, setBalance] = useState<string>("");
  const [toAddress, setToAddress] = useState<string>(
    "thor1z5wcwjs0jymr589zns48n5l0x6n2twm7dfkelp"
  );
  const [amount, setAmount] = useState<string>("1000000");
  const [passwd, setPasswd] = useState<string>("2vault");

  const getThorAddress = async () => {
    if (!vaultPublicKeyEcdsa) return;

    const vaultInfo = await (
      await getVault(vaultPublicKeyEcdsa, passwd)
    ).json();
    const walletCore = await initWasm();
    const thorPublicKeyStr = await getDerivedPublicKey(
      vaultPublicKeyEcdsa,
      vaultInfo.hex_chain_code,
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
    return {
      thorPublicKeyStr,
      thorAddress,
      vaultLocalPartyId: vaultInfo.local_party_id,
      vaultHexChainCode: vaultInfo.hex_chain_code,
    };
  };

  const getBalance = async () => {
    if (!vaultPublicKeyEcdsa) return;

    try {
      const resp = await getThorAddress();
      if (!resp || !resp.thorAddress) return;
      const data = await (await getBalances(resp.thorAddress)).json();
      const runeBalance = data.balances.find(
        (bal: any) => bal.denom === "rune"
      );
      setBalance(runeBalance ? runeBalance.amount : "0");
    } catch (error) {
      console.error("Error fetching balance:", error);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-md w-96 text-black">
      <h1 className="text-xl font-bold mb-4">Vault Token Send Demo</h1>
      <div className="mb-4">
        <input
          type="text"
          value={vaultPublicKeyEcdsa}
          onChange={(e) => setVaultPublicKeyEcdsa(e.target.value)}
          placeholder="Vault PublicKey Ecdsa"
          className="mt-1 p-2 w-full border border-gray-300 rounded-md mb-4"
        />
        <input
          type="password"
          value={passwd}
          onChange={(e) => setPasswd(e.target.value)}
          placeholder="Password"
          className="mt-1 p-2 w-full border border-gray-300 rounded-md"
        />
      </div>
      <button
        onClick={getBalance}
        className="w-full bg-blue-500 text-white py-2 rounded-md mb-4 hover:bg-blue-600"
      >
        Get Balance
      </button>
      {balance !== "" && (
        <div className="mb-4">
          <label className="block text-gray-700">Balance: {balance} RUNE</label>
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
        onClick={async () => {
          const resp = await getThorAddress();
          if (!resp || !resp.thorAddress) return;
          sendTransaction(
            vaultPublicKeyEcdsa,
            resp.vaultLocalPartyId,
            resp.vaultHexChainCode,
            resp.thorPublicKeyStr,
            resp.thorAddress,
            toAddress,
            amount,
            passwd
          );
        }}
        className="w-full bg-green-500 text-white py-2 rounded-md hover:bg-green-600"
      >
        Send
      </button>
    </div>
  );
};

export default StepOne;
