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
