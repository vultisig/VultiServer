import PolicyForm from "../policy-form/PolicyForm";
import { PolicyProvider } from "../../context/PolicyProvider";
import PolicyTable from "../policy-table/PolicyTable";

const Policy = () => {
  return (
    <PolicyProvider>
      <div className="left-section">
        <PolicyForm />
      </div>
      <div className="right-section">
        <PolicyTable />
      </div>
    </PolicyProvider>
  );
};

export default Policy;
