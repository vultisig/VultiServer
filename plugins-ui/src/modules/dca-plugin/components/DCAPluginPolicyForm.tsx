import { supportedTokens } from "@/modules/dca-plugin/data/tokens";
import TokenSelectorWithPriceEvaluation from "@/modules/shared/token-selector-with-price-evaluation/TokenSelectorWithPriceEvaluation";
import { Controller, SubmitHandler, useForm } from "react-hook-form";
import { Frequency, Policy } from "@/modules/dca-plugin/models/policy";
import {
  generateMinTimeInputValidation,
  orders_validation,
} from "@/modules/dca-plugin/utils/inputSpecifications";
import SwapIcon from "@/assets/Swap.svg?react";
import {
  generatePolicy,
  setDefaultFormValues,
} from "@/modules/dca-plugin/utils/policy.utils";
import DCAService from "@/modules/dca-plugin/services/dcaService";
import { useNavigate } from "react-router-dom";
import Button from "@/modules/core/components/ui/button/Button";
import SelectorTriggerButton from "@/modules/core/components/ui/select-modal/SelectorTriggerButton";
import { Input } from "@/modules/core/components/ui/input/Input";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import Form from "@/modules/core/components/ui/form/Form";
import "./DCAPolicy.css";

type DCAPolicyForm = {
  allocateAsset: string;
  buyAsset: string;
  orders: string;
  amount: string;
  interval: string;
  frequency: Frequency;
};

type DCAPluginPolicyProps = {
  data?: Policy;
  onSubmitCallback?: (data: Policy) => void;
};

const DCAPluginPolicyForm = ({
  data,
  onSubmitCallback,
}: DCAPluginPolicyProps) => {
  const defaultValues: DCAPolicyForm = setDefaultFormValues(data);
  const navigate = useNavigate();
  const { reset } = useForm<DCAPolicyForm>({
    defaultValues,
  });

  const handleUserSubmit: SubmitHandler<DCAPolicyForm> = async (submitData) => {
    const policy: Policy = generatePolicy(submitData, data);
    // check if form has passed data, this means we are editing policy
    if (data) {
      try {
        await DCAService.updatePolicy(policy);

        if (onSubmitCallback) {
          onSubmitCallback(data);
        }

        reset();
      } catch (error: any) {
        console.error("Failed to create policy:", error.message);
      }

      return;
    }

    try {
      await DCAService.createPolicy(policy);
      navigate("/dca-plugin");
      reset();
    } catch (error: any) {
      console.error("Failed to create policy:", error.message);
    }
  };

  return (
    <Form<DCAPolicyForm>
      defaultValues={defaultValues}
      onSubmit={handleUserSubmit}
      render={({ watch, control }) => (
        <>
          <h2 className="form-title">DCA Plugin Policy</h2>
          <p className="form-subtitle">
            Set up configuration settings for DCA Plugin Policy
          </p>
          <TokenSelectorWithPriceEvaluation
            fieldNames={{ inputName: "amount", selectorName: "allocateAsset" }}
            defaultAsset={defaultValues.allocateAsset}
            assets={supportedTokens}
          />
          <Button
            type="button"
            className="swap-btn"
            styleType="secondary"
            size="small"
            style={{
              backgroundColor: "#1F2A37",
              borderRadius: "8px",
              padding: "8px",
            }}
            onClick={() => console.log("todo call some function here")}
          >
            <SwapIcon width="20px" height="20px" />
          </Button>
          <div
            className="input-field-inline"
            style={{
              flexDirection: "column",
              alignItems: "flex-start",
              color: "#FFFFFF",
            }}
          >
            <label>To Buy</label>
            <Controller
              name="buyAsset"
              control={control}
              defaultValue={defaultValues.buyAsset}
              render={({ field }) => (
                <SelectorTriggerButton
                  modalHeader="Select a token"
                  placeholder="Search by token"
                  data={Object.values(supportedTokens)}
                  value={field.value}
                  onChange={field.onChange}
                >
                  {supportedTokens[field.value].image}&nbsp;
                  {supportedTokens[field.value].name}
                </SelectorTriggerButton>
              )}
            />
          </div>
          <div className="display-flex m-t-b-24 input-field-outline">
            <div className="input-container">
              <Input
                name="interval"
                {...generateMinTimeInputValidation(watch("frequency"))}
              />
              <SelectBox
                name="frequency"
                options={["minute", "hour", "day", "week", "month"]}
                defaultValue={defaultValues.frequency}
              />
            </div>

            <div className="input-field-outline">
              <div className="input-container">
                <Input name="orders" {...orders_validation} />
                <aside className="absolute">orders</aside>
              </div>
            </div>
          </div>
          <Button
            type="submit"
            styleType="primary"
            size="medium"
            className="submit"
            style={{ borderRadius: "8px" }}
          >
            {data ? "Save changes" : "Save"}
          </Button>
        </>
      )}
    />
  );
};

export default DCAPluginPolicyForm;
