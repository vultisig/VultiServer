import "./Home.css";
import PolicyForm from "./modules/policy-form/components/PolicyForm";
import { PolicyProvider } from "./modules/policy-form/context/PolicyProvider";
import PolicyTable from "./modules/policy-table/components/PolicyTable";
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
