import Button from "@/modules/core/components/ui/button/Button";
import TrashIcon from "@/assets/Trash.svg?react";
import PenIcon from "@/assets/Pen.svg?react";
import PauseIcon from "@/assets/Pause.svg?react";
import { usePolicies } from "@/modules/policy/context/PolicyProvider";
import Modal from "@/modules/core/components/ui/modal/Modal";
import PolicyForm from "@/modules/policy/components/policy-form/PolicyForm";
import { useState } from "react";

type PolicyActionsProps = {
  policyId: string;
};

const PolicyActions = ({ policyId }: PolicyActionsProps) => {
  const [modalId, setModalId] = useState("");
  const { policyMap, removePolicy } = usePolicies();

  return (
    <>
      <div style={{ display: "flex" }}>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => console.log("todo implement", policyId)}
        >
          <PauseIcon width="20px" height="20px" color="#F0F4FC" />
        </Button>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => setModalId(policyId)}
        >
          <PenIcon width="20px" height="20px" color="#F0F4FC" />
        </Button>
        <Button
          type="button"
          styleType="tertiary"
          size="small"
          style={{ color: "#DA2E2E", padding: "5px", margin: "0 5px" }}
          onClick={() => removePolicy(policyId)}
        >
          <TrashIcon width="20px" height="20px" color="#FF5C5C" />
        </Button>
      </div>

      <Modal
        isOpen={modalId !== ""}
        onClose={() => setModalId("")}
        variant="panel"
      >
        <PolicyForm
          data={policyMap.get(modalId)}
          onSubmitCallback={() => setModalId("")}
        />
      </Modal>
    </>
  );
};

export default PolicyActions;
