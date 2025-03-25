// more on the exposed methods here: https://github.com/vultisig/vultisig-windows/blob/main/clients/extension/docs/integration-guide.md

interface ProviderError {
  code: number;
  message: string;
}

const VulticonnectWalletService = {
  connectToVultiConnect: async () => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const accounts = await window.vultisig.ethereum.request({
        method: "eth_requestAccounts",
      });

      return accounts;
    } catch (error) {
      const { code, message } = error as ProviderError;
      console.error(`Connection failed - Code: ${code}, Message: ${message}`);
      throw error;
    }
  },

  getConnectedEthAccounts: async () => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const accounts = await window.vultisig.ethereum.request({
        method: "eth_accounts",
      });

      return accounts;
    } catch (error) {
      const { code, message } = error as ProviderError;
      console.error(
        `Failed to get accounts - Code: ${code}, Message: ${message}`
      );
      throw error;
    }
  },

  signCustomMessage: async (hexMessage: string, walletAddress: string) => {
    if (!window.vultisig?.ethereum) {
      alert(`No ethereum provider found. Please install VultiConnect.`);
      return;
    }

    try {
      const signature = await window.vultisig.ethereum.request({
        method: "personal_sign",
        params: [hexMessage, walletAddress],
      });

      if (signature && signature.error) {
        throw signature.error;
      }
      return signature;
    } catch (error) {
      console.error("Failed to sign the message", error);
      throw new Error("Failed to sign the message");
    }
  },

  getVaults: async () => {
    if (!window.vultisig) {
      alert(`No wallet found. Please install VultiConnect.`);
      return;
    }

    try {
      const vaults = await window.vultisig.getVaults();

      if (!vaults || vaults.length === 0) {
        throw new Error("No vaults found");
      }

      return vaults;
    } catch (error) {
      console.error(`Failed to get vaults ${error}`);
      throw error;
    }
  },
};

export default VulticonnectWalletService;
