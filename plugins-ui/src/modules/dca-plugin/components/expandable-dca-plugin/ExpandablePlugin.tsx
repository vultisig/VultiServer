import "./ExpandablePlugin.css";
import dcaPlugin from "@/assets/DCA-image.png";
import PenIcon from "@/assets/Pen.svg?react";
import TrashIcon from "@/assets/Trash.svg?react";
import Accordion from "@/modules/core/components/ui/accordion/Accordion";
import Button from "@/modules/core/components/ui/button/Button";
import { useEffect, useState } from "react";
import Modal from "@/modules/core/components/ui/modal/Modal";
import { supportedTokens } from "../../data/tokens";
import PolicyService from "@/modules/policy-form/services/policyService";
import PolicyForm from "@/modules/policy-form/components/PolicyForm";
import { PluginPolicy } from "@/modules/policy-form/models/policy";

const ExpandableDCAPlugin = () => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());

  const [modalId, setModalId] = useState("");

  useEffect(() => {
    async function fetchPolicies() {
      try {
        const fetchedPolicies = await PolicyService.getPolicies();

        const constructPolicyMap: Map<string, PluginPolicy> = new Map(
          fetchedPolicies?.map((p: PluginPolicy) => [p.id, p]) // Convert the array into [key, value] pairs
        );

        setPolicyMap(constructPolicyMap);
      } catch (error: any) {
        console.error("Failed to get policies:", error.message);
      }
    }

    fetchPolicies();
  }, []);

  const handleFormSubmit = (data: PluginPolicy) => {
    policyMap.set(data.id, data);
    setModalId("");
  };

  const deletePolicy = async (policyId: string) => {
    try {
      await PolicyService.deletePolicy(policyId);
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
                Ethereum
              </h4>
            </div>
          </>
        }
        expandButton={{ text: "See all policies", style: { color: "#33E6BF" } }}
      >
        {policyMap.size > 0 &&
          Array.from(policyMap).map(([key, value]) => (
            <div key={key} className="policy">
              <div className="group">
                {supportedTokens[value.policy.source_token_id as string]
                  ?.image || "?"}
                {supportedTokens[value.policy.destination_token_id as string]
                  ?.image || "?"}
                &nbsp;
                {supportedTokens[value.policy.source_token_id as string]
                  ?.name ||
                  `Unknown token address: ${value.policy.source_token_id}`}
                /
                {supportedTokens[value.policy.destination_token_id as string]
                  ?.name ||
                  `Unknown token address: ${value.policy.destination_token_id}`}
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
      <Modal
        isOpen={modalId !== ""}
        onClose={() => setModalId("")}
        variant="panel"
      >
        <PolicyForm
          data={policyMap.get(modalId)}
          onSubmitCallback={handleFormSubmit}
        />
      </Modal>
    </>
  );
};

export default ExpandableDCAPlugin;
