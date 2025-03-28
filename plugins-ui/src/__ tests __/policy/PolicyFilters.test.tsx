import PolicyFilters from "@/modules/policy/components/policy-filters/PolicyFilters";
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

describe("PolicyFilters", () => {
  it("should render closed select with default option All", () => {
    const mockOnChange = vi.fn();

    render(<PolicyFilters onFiltersChange={mockOnChange} />);

    expect(screen.queryByRole("list")).not.toBeInTheDocument();
    expect(screen.getByText("all")).toBeInTheDocument();
  });

  it("should open dropdown upon click", () => {
    const mockOnChange = vi.fn();

    render(<PolicyFilters onFiltersChange={mockOnChange} />);

    expect(screen.queryByRole("list")).not.toBeInTheDocument();
    const defaultOption = screen.getByText("all");

    fireEvent.click(defaultOption);
    expect(screen.queryByRole("list")).toBeInTheDocument();
  });

  it("should change filters & close dropdown upon click", () => {
    const mockOnChange = vi.fn();

    render(<PolicyFilters onFiltersChange={mockOnChange} />);

    const defaultOption = screen.getByText("all");

    fireEvent.click(defaultOption);
    const activeOption = screen.getByText("active");

    expect(activeOption).toBeInTheDocument();
    fireEvent.click(activeOption);

    expect(screen.queryByRole("list")).not.toBeInTheDocument();
    expect(mockOnChange).toBeCalledWith([
      {
        id: "status",
        value: true,
      },
    ]);
  });
});
