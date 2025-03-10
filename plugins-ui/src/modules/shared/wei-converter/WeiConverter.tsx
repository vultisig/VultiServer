import React, { useEffect, useState } from "react";
import { WidgetProps } from "@rjsf/utils";
import { ethers } from "ethers";
import { supportedTokens } from "../data/tokens";
import "./WeiConverter.css";

const DEBOUNCE_DELAY = 500;

const WeiConverter: React.FC<WidgetProps> = ({
  value,
  onChange,
  formContext,
  schema,
}) => {
  const selectedToken = formContext?.sourceTokenId;
  const [inputValue, setInputValue] = useState("");

  useEffect(() => {
    if (!value || !schema.pattern) {
      setInputValue("");
      return;
    }

    let decimals = supportedTokens[selectedToken]?.decimals;
    try {
      const formattedValue = ethers.formatUnits(value, decimals);
      setInputValue(formattedValue);
    } catch (error) {
      console.error(error);
    }
  }, [value, selectedToken]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (!schema.pattern) {
        onChange("");
        return;
      }

      let decimals = supportedTokens[selectedToken]?.decimals;
      try {
        const convertedValue = ethers
          .parseUnits(inputValue, decimals)
          .toString();
        onChange(convertedValue);
      } catch (error) {
        console.error(error);
      }
    }, DEBOUNCE_DELAY);

    return () => clearTimeout(timeout);
  }, [inputValue, selectedToken]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  return (
    <>
      <input
        id="wei"
        type="number"
        value={inputValue}
        onChange={handleChange}
        data-testid="wei-converter"
      />
    </>
  );
};

export default WeiConverter;
