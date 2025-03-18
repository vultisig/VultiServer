import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import Form from "@rjsf/core";
import validator from "@rjsf/validator-ajv8";
import WeiConverter from "@/modules/shared/wei-converter/WeiConverter";
import { RJSFSchema } from "@rjsf/utils";
import { WETH_TOKEN } from "@/modules/shared/data/tokens";

const schema: RJSFSchema = {
  type: "object",
  properties: {
    customInput: {
      type: "string",
      title: "Wei converter widget",
      pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
    },
  },
};

const uiSchema = {
  customInput: { "ui:widget": WeiConverter },
};

describe("WeiConverter", () => {
  it("should render with correct value converted from wei to token value", () => {
    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000" }}
        formContext={{
          sourceTokenId: WETH_TOKEN,
        }}
      />
    );

    const input = screen.getByTestId("wei-converter");
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute("value", "1.0");
  });

  it("should not convert values that are not matching the pattern", () => {
    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000a" }}
        formContext={{
          sourceTokenId: WETH_TOKEN,
        }}
      />
    );

    const input = screen.getByTestId("wei-converter");
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute("value", "");
  });

  it("should not convert values if pattern is missing", () => {
    const schema: RJSFSchema = {
      type: "object",
      properties: {
        customInput: {
          type: "string",
          title: "Wei converter widget",
        },
      },
    };

    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000" }}
        formContext={{
          sourceTokenId: WETH_TOKEN,
        }}
      />
    );

    const input = screen.getByTestId("wei-converter");
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute("value", "");
  });

  it("should error handle ethers.formatUnits properly if pattern allows letters ", () => {
    const schema: RJSFSchema = {
      type: "object",
      properties: {
        customInput: {
          type: "string",
          title: "Wei converter widget",
          pattern: "^[a-z0-9]+$",
        },
      },
    };

    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000a" }}
        formContext={{
          sourceTokenId: WETH_TOKEN,
        }}
      />
    );

    expect(consoleSpy).toHaveBeenCalledWith(expect.any(TypeError));

    consoleSpy.mockRestore();
  });

  it("should handle ethers.formatUnits properly if token from context is missing in the supported token list ", () => {
    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000" }}
        formContext={{
          sourceTokenId: "missing_token", // this won't throw error, the defualt is to turn it into 18 decimalsa
        }}
      />
    );

    const input = screen.getByTestId("wei-converter");
    expect(input).toHaveAttribute("value", "1.0");
  });

  it("should call onChange when user types with the amount converted to wei", async () => {
    const mockOnChange = vi.fn();

    render(
      <Form
        schema={schema}
        uiSchema={uiSchema}
        validator={validator}
        formData={{ customInput: "1000000000000000000" }}
        onChange={(e) => mockOnChange(e.formData.customInput)}
        formContext={{
          sourceTokenId: WETH_TOKEN,
        }}
      />
    );

    const input = screen.getByTestId("wei-converter");

    fireEvent.change(input, { target: { value: "0.15" } });

    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith("150000000000000000");
    });
  });
});
