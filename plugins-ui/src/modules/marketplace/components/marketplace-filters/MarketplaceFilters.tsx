import Button from "@/modules/core/components/ui/button/Button";
import Grid from "@/assets/Grid.svg?react";
import List from "@/assets/List.svg?react";
import "./MarketplaceFilters.css";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";

type MarketplaceFiltersProps = {
  viewFilter: ViewFilter;
  onChange: (view: ViewFilter) => void;
};

const MarketplaceFilters = ({
  viewFilter,
  onChange,
}: MarketplaceFiltersProps) => {
  const [view, setView] = useState<ViewFilter>(viewFilter);

  const changeView = (view: ViewFilter) => {
    setView(view);
    onChange(view);
  };

  return (
    <div className="filters">
      <Button
        ariaLabel="Grid view"
        type="button"
        styleType="tertiary"
        size="medium"
        className={`view-filter ${view === "grid" ? "active" : ""}`}
        onClick={() => changeView("grid")}
      >
        <Grid width="20px" height="20px" color="#F0F4FC" />
      </Button>
      <Button
        ariaLabel="List view"
        type="button"
        styleType="tertiary"
        size="medium"
        className={`view-filter ${view === "list" ? "active" : ""}`}
        onClick={() => changeView("list")}
      >
        <List width="20px" height="20px" color="#F0F4FC" />
      </Button>
    </div>
  );
};

export default MarketplaceFilters;
