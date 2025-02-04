import { Controller, useFormContext, useWatch } from "react-hook-form";
import { Input } from "../../core/components/ui/input/Input";
import { allocate_from_validation } from "@/modules/dca-plugin/utils/inputSpecifications";
import SelectorTriggerButton from "../../core/components/ui/select-modal/SelectorTriggerButton";
import { useEffect, useState } from "react";
import { TokenPrices } from "@/modules/dca-plugin/models/token";
import TokenPricesService from "@/modules/dca-plugin/services/tokenPricesService";
import { getTokenPrice } from "@/modules/dca-plugin/utils/token.util";
import { Option } from "@/modules/core/models/select-modal";

type TokenSelectorWithPriceEvaluationProps = {
  fieldNames: { inputName: string; selectorName: string };
  defaultAsset: string;
  assets: { [key: string]: Option };
};

const TokenSelectorWithPriceEvaluation = ({
  fieldNames,
  defaultAsset,
  assets,
}: TokenSelectorWithPriceEvaluationProps) => {
  const { control } = useFormContext();
  const [dollarCost, setDollarCost] = useState<string>("0");
  const amount = useWatch({ name: fieldNames.inputName });

  useEffect(() => {
    onTokenToAllocateChange(defaultAsset);
  }, []);

  const onTokenToAllocateChange = async (tokenAddress: string) => {
    try {
      const response: TokenPrices =
        await TokenPricesService.getTokenPricesByAddress(tokenAddress);
      const tokenPrice = getTokenPrice(response, tokenAddress);

      setDollarCost(tokenPrice);
    } catch (error: any) {
      console.error(
        `Failed to get token by address ${tokenAddress}:`,
        error.message
      );
    }
  };

  return (
    <div className="input-field-inline">
      <div>
        <Input name={fieldNames.inputName} {...allocate_from_validation} />
        <small className="dollar-equivalent">$ {+dollarCost * amount}</small>
      </div>
      <Controller
        name={fieldNames.selectorName}
        control={control}
        defaultValue={defaultAsset}
        render={({ field }) => (
          <SelectorTriggerButton
            modalHeader="Select a token"
            placeholder="Search by token"
            data={Object.values(assets)}
            value={field.value}
            onChange={(value) => {
              field.onChange(value);
              onTokenToAllocateChange(value);
            }}
          >
            {assets[field.value].image}&nbsp;
            {assets[field.value].name}
          </SelectorTriggerButton>
        )}
      />
    </div>
  );
};

export default TokenSelectorWithPriceEvaluation;
