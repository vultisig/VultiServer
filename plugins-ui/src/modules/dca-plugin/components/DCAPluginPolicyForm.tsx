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
  chain: string;
  provider: any;
  data?: Policy;
  onSubmitCallback?: (data: Policy) => void;
};

const DCAPluginPolicyForm = ({
  chain,
  provider,
  data,
  onSubmitCallback,
}: DCAPluginPolicyProps) => {
  const defaultValues: DCAPolicyForm = setDefaultFormValues(data);
  const navigate = useNavigate();
  const { reset } = useForm<DCAPolicyForm>({
    defaultValues,
  });

  const signCustomMessage = async (hexMessage: string, walletAddress: string) => {
      if (window.vultisig?.ethereum) {
          try {
              const signature = await window.vultisig.ethereum.request({
                  method: "personal_sign",
                  params: [hexMessage, walletAddress],
              });
              return signature
          } catch (error) {
              return Error("Failed to sign the message: " + error);
          }
      }
  };

  const getConnectedEthereum = async (provider: any) => {
      if (provider) {
          try {
              const accounts = await provider.request({ method: "eth_accounts" });
              if (accounts.length) {
                  console.log(`Currently connected address:`, accounts);
                  return accounts;
              } else {
                  console.log(`Currently no account is connected to this dapp:`, accounts);
              }
          } catch (error) {
              console.error("Ethereum getting connected accounts failed", error);
          }
      } else {
          alert("No Ethereum provider found. Please install VultiConnect or MetaMask.");
      }
  };

  const getConnectedAccountsChain = async (chain: string, provider: any) => {
      if (provider) {
          try {
              const accounts = await provider.request({ method: "get_accounts" });
              if (accounts.length) {
                  console.log(`Currently connected address:`, accounts);
                  return accounts;
              } else {
                  console.log(`Currently no account is connected to this dapp:`, accounts);
              }
          } catch (error) {
              console.error(`${chain} getting connected accounts failed`, error);
          }
      } else {
          alert(`No ${chain} provider found. Please install VultiConnect.`);
      }
  };
            
  const signPolicy = async (policy: Policy, chain: string, provider: any) => {
      const toHex = (str: string) => {
          return Array
              .from(str)
              .map(char => char.charCodeAt(0).toString(16).padStart(2, '0'))
              .join('');
      }  

      const serializedPolicy = JSON.stringify(policy);
      const hexMessage = toHex(serializedPolicy);
      
      let accounts = [];
      if (chain === "ethereum") {
          accounts = await getConnectedEthereum(provider);
      } else {
          accounts = await getConnectedAccountsChain(chain, provider);
      }

      if (!accounts || accounts.length === 0) {
          console.error('Need to connect to an Ethereum wallet');
          return;
      }

      const signature = await signCustomMessage(hexMessage, accounts[0]);
      if (signature == null || signature instanceof Error) {
          // TODO: show propper error message to the user
          console.error("Failed to sign the message");
          return;
      }
      // if the popup gets closed during key singing (even if threshold is reached)
      // it will terminate and not generate a signature
      policy.signature = signature;
  }   

  const handleUserSubmit: SubmitHandler<DCAPolicyForm> = async (submitData) => {
    const policy: Policy = generatePolicy(submitData, data);
    // check if form has passed data, this means we are editing policy
    if (data) {
      try {
        await signPolicy(policy, chain, provider);
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
      await signPolicy(policy, chain, provider);
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
