import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import { ColumnFiltersState } from "@tanstack/react-table";

type TableFitlersProps = {
  onFiltersChange: (filters: ColumnFiltersState) => void;
};

const statusMap: Record<Status, boolean | null> = {
  all: null,
  active: true,
  inactive: false,
};

type Status = "all" | "active" | "inactive";
const policyStatuses = ["all", "active", "inactive"];
const DEFAULT_STATUS = "all";

const isStatus = (value: string): value is Status => {
  return policyStatuses.includes(value);
};

// if needed this function could be further extended or refactor in case not all policies have active status
const PolicyFilters = ({ onFiltersChange }: TableFitlersProps) => {
  const handleStatusChange = (change: string) => {
    if (!isStatus(change)) {
      onFiltersChange([]);
      return;
    }

    const value = statusMap[change];
    onFiltersChange(value === null ? [] : [{ id: "status", value }]);
  };

  return (
    <SelectBox
      label="See:"
      options={policyStatuses}
      value={DEFAULT_STATUS}
      onSelectChange={handleStatusChange}
      style={{ width: "20%" }}
    />
  );
};

export default PolicyFilters;
