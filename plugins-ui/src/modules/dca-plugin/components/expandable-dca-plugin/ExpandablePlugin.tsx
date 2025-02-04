import "./ExpandablePlugin.css";
import dcaPlugin from "@/assets/DCA-image.png";
import PenIcon from "@/assets/Pen.svg?react";
import TrashIcon from "@/assets/Trash.svg?react";
import Accordion from "@/modules/core/components/ui/accordion/Accordion";
import Button from "@/modules/core/components/ui/button/Button";
import { useEffect, useState } from "react";
import DCAService from "../../services/dcaService";
import { Policy } from "../../models/policy";
import { Link } from "react-router-dom";
import Modal from "@/modules/core/components/ui/modal/Modal";
import DCAPluginPolicyForm from "../DCAPluginPolicyForm";
import { supportedTokens } from "../../data/tokens";

const ExpandableDCAPlugin = () => {
  const [policyMap, setPolicyMap] = useState(new Map<string, Policy>());
  const [modalId, setModalId] = useState("");

  useEffect(() => {
    async function fetchPolicies() {
      try {
        const fetchedPolicies = await DCAService.getPolicies();

        const constructPolicyMap: Map<string, Policy> = new Map(
          fetchedPolicies?.map((p: Policy) => [p.id, p]) // Convert the array into [key, value] pairs
        );

        setPolicyMap(constructPolicyMap);
      } catch (error: any) {
        console.error("Failed to get policies:", error.message);
      }
    }

    fetchPolicies();
  }, []);

  const handleFormSubmit = (data: Policy) => {
    policyMap.set(data.id, data);
    setModalId("");
  };

  const deletePolicy = async (policyId: string) => {
    try {
      await DCAService.deletePolicy(policyId);
      const updatedPolicyMap = new Map(policyMap);
      updatedPolicyMap.delete(policyId);
      setPolicyMap(updatedPolicyMap);
    } catch (error: any) {
      console.error("Failed to delete policy:", error.message);
    }
  };

  return (
    <>
      <Accordion
        header={
          <>
            <img src={dcaPlugin} alt="" width="72px" height="72px" />
            <div className="headers">
              <div className="status">Active</div>
              <h3>DCA Plugin</h3>
              <h4>
                Allows you to dollar cost average into any supported token like
                Etherium
              </h4>
            </div>
            <Link to="/dca-plugin/form">
              <Button
                type="button"
                styleType="primary"
                size="medium"
                style={{ paddingTop: 8, paddingBottom: 8 }}
              >
                Add new
              </Button>
            </Link>
          </>
        }
        expandButton={{ text: "See all policies", style: { color: "#33E6BF" } }}
      >
        {policyMap.size > 0 &&
          Array.from(policyMap).map(([key, value]) => (
            <div key={key} className="policy">
              <div className="group">
                {supportedTokens[value.policy.source_token_id].image}
                {supportedTokens[value.policy.destination_token_id].image}&nbsp;
                {supportedTokens[value.policy.source_token_id].name}/
                {supportedTokens[value.policy.destination_token_id].name}
                {key}
              </div>
              <div className="group">
                <Button
                  type="button"
                  styleType="tertiary"
                  size="small"
                  style={{ color: "#33E6BF" }}
                  onClick={() => setModalId(key)}
                >
                  <PenIcon />
                  Edit
                </Button>
                <Button
                  type="button"
                  styleType="tertiary"
                  size="small"
                  style={{ color: "#DA2E2E" }}
                  onClick={() => deletePolicy(key)}
                >
                  <TrashIcon width="20px" height="20px" />
                  Remove
                </Button>
              </div>
            </div>
          ))}
        {policyMap.size === 0 && <>There is nothing to show yet.</>}
      </Accordion>
      ;
      <Modal isOpen={modalId !== ""} onClose={() => setModalId("")}>
        <DCAPluginPolicyForm
          data={policyMap.get(modalId)}
          onSubmitCallback={handleFormSubmit}
        />
      </Modal>
    </>
  );
};

export default ExpandableDCAPlugin;
