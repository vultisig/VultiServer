import Button from "@/modules/core/components/ui/button/Button";
import VulticonnectWalletService from "./vulticonnectWalletService";
import { useState } from "react";
import PolicyService from "@/modules/policy/services/policyService";
import { derivePathMap, getHexMessage } from "./wallet.utils";

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
      case "ethereum": {
        const accounts =
          await VulticonnectWalletService.connectToVultiConnect();
        if (accounts.length && accounts[0]) {
          setConnectedWallet(true);
        }

        const vaults = await VulticonnectWalletService.getVaults();

        const publicKey = vaults[0].publicKeyEcdsa;
        if (publicKey) {
          localStorage.setItem("publicKey", publicKey);
        }
        const chainCodeHex = vaults[0].hexChainCode;
        const derivePath = derivePathMap[chain];

        const hexMessage = getHexMessage(publicKey);

        const signature = await VulticonnectWalletService.signCustomMessage(
          hexMessage,
          accounts[0]
        );

        if (signature && typeof signature === "string") {
          const token = await PolicyService.getAuthToken(
            hexMessage,
            signature,
            publicKey,
            chainCodeHex,
            derivePath
          );
          localStorage.setItem("authToken", token);
        }

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
