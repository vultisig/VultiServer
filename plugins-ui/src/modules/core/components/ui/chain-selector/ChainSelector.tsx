const networks = [
  { id: "bitcoin", name: "Bitcoin" },
  { id: "ethereum", name: "Ethereum" },
  { id: "thorchain", name: "Thorchain" },
  // TODO: add more networks as needed
];

type ChainSelectorProps = {
  chain: string;
  setChain: (chain: string) => void;
};

const ChainSelector = ({ chain, setChain }: ChainSelectorProps) => {
  return (
    <select
      style={{
        padding: "10px",
        borderRadius: "10px",
        background: "#abc",
        position: "relative",
      }}
      value={chain}
      onChange={(e) => setChain(e.target.value)}
    >
      {networks.map((network) => (
        <option key={network.id} value={network.id}>
          {network.name}
        </option>
      ))}
    </select>
  );
};

export default ChainSelector;
