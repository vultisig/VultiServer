import "./Home.css";
import PolicyForm from "./modules/policy/components/policy-form/PolicyForm";
import { PolicyProvider } from "./modules/policy/context/PolicyProvider";
import PolicyTable from "./modules/policy/components/policy-table/PolicyTable";
import Wallet from "./modules/shared/wallet/Wallet";

const Home = () => {
  return (
    <div className="container">
      <div className="navbar">
        <Wallet />
      </div>
      <div className="content">
        <PolicyProvider>
          <div className="left-section">
            <PolicyForm />
          </div>
          <div className="right-section">
            <PolicyTable />
          </div>
        </PolicyProvider>
      </div>
    </div>
  );
};

export default Home;
