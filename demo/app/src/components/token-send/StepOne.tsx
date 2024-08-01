interface StepOneProps {
  vaultPublicKeyEcdsa: string;
  balance: string;
  toAddress: string;
  amount: string;
  passwd: string;
  setVaultPublicKeyEcdsa: (value: string) => void;
  setPasswd: (value: string) => void;
  setToAddress: (value: string) => void;
  setAmount: (value: string) => void;
  getBalance: () => void;
  sendTransaction: () => void;
}

const StepOne = ({
  vaultPublicKeyEcdsa,
  balance,
  toAddress,
  amount,
  passwd,
  setVaultPublicKeyEcdsa,
  setToAddress,
  setAmount,
  setPasswd,
  getBalance,
  sendTransaction,
}: StepOneProps) => {
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
      {!balance && (
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
        onClick={sendTransaction}
        className="w-full bg-green-500 text-white py-2 rounded-md hover:bg-green-600"
      >
        Send
      </button>
    </div>
  );
};

export default StepOne;
