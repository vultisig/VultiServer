import { ethers } from "ethers";
import { supportedTokens } from "../data/tokens";
import { ChangeEvent, useEffect, useState } from "react";
import "./WeiConverter.css";
import { RJSFSchema, Widget } from "@rjsf/utils";

interface CustomFormContext {
  sourceTokenId: string;
}

const WeiConverter: Widget<RJSFSchema, any, CustomFormContext> = (props) => {
  let fromWei = "";

  const regex = new RegExp(props.schema.pattern);

  if (
    props.value &&
    props.registry.formContext.sourceTokenId &&
    regex.test(props.value)
  ) {
    fromWei = ethers.formatUnits(
      props.value,
      supportedTokens[props.registry.formContext.sourceTokenId].decimals
    );
  }
  const [inputValue, setInputValue] = useState(fromWei);
  const [_, setDebouncedInputValue] = useState("");

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      setDebouncedInputValue(inputValue);
      const regex = new RegExp(props.schema.pattern);

      if (!regex.test(inputValue)) {
        props.onChange(inputValue);
      }

      if (
        regex.test(inputValue) &&
        inputValue &&
        props.registry.formContext.sourceTokenId
      ) {
        const amountToWei = ethers
          .parseUnits(
            inputValue,
            supportedTokens[props.registry.formContext.sourceTokenId].decimals
          )
          .toString();

        props.onChange(amountToWei);
      }
    }, 200);
    return () => clearTimeout(timeoutId);
  }, [inputValue, 200]);

  const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
    setInputValue(event.target.value);
  };

  return (
    <>
      <input
        id="wei"
        type="text"
        value={inputValue}
        onChange={handleInputChange}
      />
    </>
  );
};

export default WeiConverter;
