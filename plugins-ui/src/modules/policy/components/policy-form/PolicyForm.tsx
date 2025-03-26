import Form, { IChangeEvent } from "@rjsf/core";
import validator from "@rjsf/validator-ajv8";
import "./PolicyForm.css";
import { generatePolicy } from "../../utils/policy.util";
import { PluginPolicy, PolicySchema } from "../../models/policy";
import { useEffect, useState } from "react";
import { usePolicies } from "../../context/PolicyProvider";
import { TitleFieldTemplate } from "../policy-title/PolicyTitle";
import TokenSelector from "@/modules/shared/token-selector/TokenSelector";
import WeiConverter from "@/modules/shared/wei-converter/WeiConverter";
import { RJSFValidationError } from "@rjsf/utils";

type PolicyFormProps = {
  data?: PluginPolicy;
  onSubmitCallback?: (data: PluginPolicy) => void;
};

const PolicyForm = ({ data, onSubmitCallback }: PolicyFormProps) => {
  const policyId = data?.id || "";

  const initialFormData = data ? data.policy : {}; // Define the initial form state
  const [formData, setFormData] = useState(initialFormData);
  const { addPolicy, updatePolicy, policySchemaMap, pluginType } =
    usePolicies();
  const [schema, setSchema] = useState<PolicySchema | null>(null);

  useEffect(() => {
    const savedSchema = policySchemaMap.get(pluginType);
    if (savedSchema) {
      setSchema(savedSchema);
    }
  }, [policySchemaMap]);

  const [formKey, setFormKey] = useState(0); // Changing this forces remount

  const onChange = (e: IChangeEvent) => {
    setFormData(e.formData);
  };

  const onSubmit = async (submitData: IChangeEvent) => {
    if (schema?.form) {
      const policy: PluginPolicy = generatePolicy(
        schema.form.plugin_version,
        schema.form.policy_version,
        schema.form.plugin_type,
        policyId,
        submitData.formData
      );

      // check if form has policyId, this means we are editing policy
      if (policyId) {
        try {
          updatePolicy(policy).then((updatedSuccessfully) => {
            if (updatedSuccessfully && onSubmitCallback) {
              onSubmitCallback(policy);
            }
          });
        } catch (error: any) {
          console.error("Failed to update policy:", error.message);
        }

        return;
      }

      try {
        addPolicy(policy).then((addedSuccessfully) => {
          if (!addedSuccessfully) return;

          setFormData(initialFormData); // Reset formData to initial state
          setFormKey((prevKey) => prevKey + 1); // Change key to force remount
          if (onSubmitCallback) {
            onSubmitCallback(policy);
          }
        });
      } catch (error: any) {
        console.error("Failed to create policy:", error.message);
      }
    }
  };

  const transformErrors = (errors: RJSFValidationError[]) => {
    return errors.map((error) => {
      if (error.name === "pattern") {
        error.message = "should be a positive number";
      }
      if (error.name === "required") {
        error.message = "required";
      }
      return error;
    });
  };

  return (
    <div className="policy-form">
      {schema && (
        <Form
          key={formKey} // Forces full re-render on reset
          idPrefix={pluginType}
          schema={schema.form.schema}
          uiSchema={schema.form.uiSchema}
          validator={validator}
          formData={formData}
          onSubmit={onSubmit}
          onChange={onChange}
          showErrorList={false}
          templates={{ TitleFieldTemplate }}
          widgets={{ TokenSelector, WeiConverter }}
          transformErrors={transformErrors}
          liveValidate={!!policyId}
          formContext={{ sourceTokenId: formData.source_token_id as string }} // sourceTokenId is needed in WeiConverter to get the rigth decimal places based on token address
        />
      )}
    </div>
  );
};

export default PolicyForm;
