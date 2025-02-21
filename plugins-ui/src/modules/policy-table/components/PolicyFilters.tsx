import { ColumnFiltersState } from "@tanstack/react-table";
import { useState } from "react";

type TableFitlersProps = {
  onFiltersChange: (filters: ColumnFiltersState) => void;
};

// if needed this function could be further extended
const PolicyFilters = ({ onFiltersChange }: TableFitlersProps) => {
  const [status, setStatus] = useState<boolean | null>(null);

  const handleStatusChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const value =
      event.target.value === "true"
        ? true
        : event.target.value === "false"
          ? false
          : null;
    setStatus(value);

    if (value || value === false) {
      onFiltersChange([{ id: "status", value: value }]);
      return;
    }

    onFiltersChange([]);
  };

  return (
    <>
      <label htmlFor="status">Status:</label>
      <select id="status" value={String(status)} onChange={handleStatusChange}>
        <option value="all">All</option>
        <option value="true">Active</option>
        <option value="false">Inactive</option>
      </select>
    </>
  );
};

export default PolicyFilters;
