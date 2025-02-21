import Form, { IChangeEvent } from "@rjsf/core";
import validator from "@rjsf/validator-ajv8";
import "./PolicyForm.css";
import {
  defaultFormData,
  getFormData,
  getSummaryData,
  getUiSchema,
  schema,
} from "../schema/dcaSchema"; // todo these should be dynamic once we have the marketplace
import { generatePolicy } from "../utils/policy.util";
import { PluginPolicy } from "../models/policy";
import { useState } from "react";
import Summary from "@/modules/shared/summary/Summary";
import { SummaryData } from "@/modules/shared/summary/summary.model";
import { usePolicies } from "../context/PolicyProvider";

type PolicyFormProps = {
  data?: PluginPolicy;
  onSubmitCallback?: (data: PluginPolicy) => void;
};

const PolicyForm = ({ data, onSubmitCallback }: PolicyFormProps) => {
  const policyId = data?.id || "";

  const initialFormData = data ? getFormData(data?.policy) : defaultFormData; // Define the initial form state
  const [formData, setFormData] = useState(initialFormData);
  const { addPolicy, updatePolicy } = usePolicies();

  const initialSummaryData: SummaryData | null =
    getSummaryData(initialFormData);
  const [summaryData, setSummaryData] = useState<SummaryData | null>(
    initialSummaryData
  );

  const [formKey, setFormKey] = useState(0); // Changing this forces remount

  const onChange = (e: IChangeEvent) => {
    setFormData(e.formData);
    setSummaryData(getSummaryData(e.formData));
  };

  const onSubmit = async (submitData: IChangeEvent) => {
    // todo do not hardcode dca once we have the marketplace
    const policy: PluginPolicy = generatePolicy(
      "dca",
      policyId,
      submitData.formData
    );

    // check if form has policyId, this means we are editing policy
    if (policyId) {
      try {
        await updatePolicy(policy);

        if (onSubmitCallback) {
          onSubmitCallback(policy);
        }
      } catch (error: any) {
        console.error("Failed to create policy:", error.message);
      }

      return;
    }

    try {
      addPolicy(policy);
      setFormData(initialFormData); // Reset formData to initial state
      setFormKey((prevKey) => prevKey + 1); // Change key to force remount
      if (onSubmitCallback) {
        onSubmitCallback(policy);
      }
    } catch (error: any) {
      console.error("Failed to create policy:", error.message);
    }
  };

  return (
    <div className="policy-form">
      <Form
        key={formKey} // Forces full re-render on reset
        idPrefix="dca" // todo this should be dynamic once we have the marketplace
        schema={schema}
        uiSchema={getUiSchema(policyId, formData)}
        validator={validator}
        formData={formData}
        onSubmit={onSubmit}
        onChange={onChange}
        showErrorList="bottom"
      />
      {summaryData && <Summary {...summaryData} />}
    </div>
  );
};

export default PolicyForm;
