import { useEffect, useState } from "react";
import { getCurrentProvider } from "./wallet.utils";
import ChainSelector from "@/modules/core/components/ui/chain-selector/ChainSelector";
import Button from "@/modules/core/components/ui/button/Button";

const Wallet = () => {
  const [chain, setChain] = useState(() => localStorage.getItem("chain"));
  const [provider, setProvider] = useState<any>(null);

  useEffect(() => {
    localStorage.setItem("chain", chain);
    const currentProvider = getCurrentProvider(chain);
    setProvider(currentProvider);
    console.log("Chain:", chain);
  }, [chain]);

  const connectEthereum = async (provider: any) => {
    if (provider) {
      try {
        const accounts = await provider.request({
          method: "eth_requestAccounts",
        });
        console.log("Connected to ethereum wallet:", accounts);
      } catch (error) {
        console.error("Ethereum connection failed", error);
      }
    } else {
      alert(
        "No ethereum provider found. Please install VultiConnect or MetaMask."
      );
    }
  };

  const connectChain = async (chain: string, provider: any) => {
    if (provider) {
      try {
        const accounts = await provider.request({ method: "request_accounts" });
        console.log(`Connected to ${chain} wallet:`, accounts);
      } catch (error) {
        console.error(`${chain} connection failed`, error);
      }
    } else {
      alert(`No ${chain} provider found. Please install VultiConnect.`);
    }
  };

  const connectWallet = async (chain: string, provider: any) => {
    if (chain === "ethereum") {
      if (window.vultisig?.ethereum) {
        console.log("VultiConnect Ethereum provider is available!");
        await connectEthereum(provider);
      } else if (window.ethereum) {
        console.log("Ethereum provider available (MetaMask or VultiConnect)");
        // Fallback to MetaMask-compatible logic
      }
    } else {
      if (window[chain] || window.vultisig?.[chain]) {
        console.log("VultiConnect provider is available!");
        await connectChain(chain, provider);
      } else {
        console.log("No compatible provider found.");
        alert(`No ${chain} provider found. Please install VultiConnect.`);
      }
    }
  };

  return (
    <>
      <ChainSelector chain={chain} setChain={setChain} />

      <Button
        size="medium"
        styleType="primary"
        type="button"
        onClick={() => connectWallet(chain, provider)}
      >
        Connect Wallet
      </Button>
    </>
  );
};

export default Wallet;
