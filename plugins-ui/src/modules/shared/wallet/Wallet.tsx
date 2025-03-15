import Button from "@/modules/core/components/ui/button/Button";
import VulticonnectWalletService from "./vulticonnectWalletService";
import { useState } from "react";

const Wallet = () => {
  let chain = localStorage.getItem("chain") as string;

  if (!chain) {
    localStorage.setItem("chain", "ethereum");
    chain = localStorage.getItem("chain") as string;
  }

  const [connectedWallet, setConnectedWallet] = useState(false);

  const connectWallet = async (chain: string) => {
    switch (chain) {
      // add more switch cases as more chains are supported
      case "ethereum":
        {
          const accounts = await VulticonnectWalletService.connectToVultiConnect();
          if (accounts.length && accounts[0]) {
            setConnectedWallet(true);
          }

          const vaults = await window.vultisig.getVaults();
          if (!vaults || vaults.length === 0) {
            throw new Error("No vaults found");
          }

          let publicKey = vaults[0].publicKeyEcdsa
          let chainCodeHex = vaults[0].hexChainCode
          const derivePath = "m/44'/60'/0'/0/0"; // Using standard Ethereum derivation path

          const messageToSign = (publicKey.startsWith('0x') ? publicKey.slice(2) : publicKey) + "1";
          const hexMessage = toHex(messageToSign);

          const signature = await VulticonnectWalletService.signCustomMessage(
              hexMessage,
              accounts[0]
          );
          
          const token = await VulticonnectWalletService.getAuthToken(
            hexMessage,
            signature,
            publicKey,
            chainCodeHex,
            derivePath
          );
          console.log(token)
          localStorage.setItem("authToken", token);
          break;
        }

      default:
        alert(`Chain ${chain} is currently not supported.`); // toast
        break;
    }
  };

  return (
    <Button
      size="medium"
      styleType="primary"
      type="button"
      onClick={() => connectWallet(chain)}
    >
      {connectedWallet ? "Connected" : "Connect Wallet"}
    </Button>
  );
};

export default Wallet;


const toHex = (str: string): string => {
  return (
      "0x" +
      Array.from(str)
          .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
          .join("")
  );
};