/// <reference types="vite-plugin-svgr/client" />

import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import DCAPluginPolicyForm from "./modules/dca-plugin/components/DCAPluginPolicyForm";
import ExpandableDCAPlugin from "./modules/dca-plugin/components/expandable-dca-plugin/ExpandablePlugin";
import { useState, useEffect } from 'react';
import ChainSelector from './modules/core/components/ui/chain-selector/ChainSelector';
import Button from './modules/core/components/ui/button/Button';

// TODO: refactor the chain selector

const App = () => {
    const [chain, setChain] = useState(() => localStorage.getItem("chain"));
    const [provider, setProvider] = useState<any>(null);

    const getCurrentProvider = (chain: string) => {
        return chain === "ethereum"
            ? window.vultisig?.ethereum || window.ethereum
            : window[chain] || window.vultisig?.[chain];
    };

    useEffect(() => {
        localStorage.setItem("chain", chain);
        const currentProvider = getCurrentProvider(chain);
        setProvider(currentProvider);
        console.log("Chain:", chain);
    }, [chain]);

    const connectEthereum = async (provider: any) => {
        if (provider) {
            try {
                const accounts = await provider.request({ method: "eth_requestAccounts" });
                console.log("Connected to ethereum wallet:", accounts);
            } catch (error) {
                console.error("Ethereum connection failed", error);
            }
        } else {
            alert("No ethereum provider found. Please install VultiConnect or MetaMask.");
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
    <BrowserRouter>
        <ChainSelector chain={chain} setChain={setChain} />

        <Button onClick={() => connectWallet(chain, provider) }>
            Connect Wallet
        </Button>

        <Routes>
            <Route path="/dca-plugin">
                <Route index element={<ExpandableDCAPlugin chain={chain} provider={provider} />} />
                <Route path="/dca-plugin/form" element={<DCAPluginPolicyForm chain={chain} provider={provider} />} />
            </Route>
        </Routes>
    </BrowserRouter>
)};

export default App;
