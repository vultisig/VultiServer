import "./Home.css";
import ExpandableDCAPlugin from "./modules/dca-plugin/components/expandable-dca-plugin/ExpandablePlugin";
import PolicyForm from "./modules/policy-form/components/PolicyForm";
import Wallet from "./modules/shared/wallet/Wallet";

const Home = () => {
  const handleFormSubmit = (data: any) => {
    // todo implement
    console.log(data);
  };
  return (
    <div className="container">
      <div className="navbar">
        <Wallet />
      </div>
      <div className="content">
        <div className="left-section">
          <PolicyForm onSubmitCallback={handleFormSubmit} />
        </div>
        <div className="right-section">
          {/* todo this will become table at one point */}
          <ExpandableDCAPlugin />
        </div>
      </div>
    </div>
  );
};

export default Home;
