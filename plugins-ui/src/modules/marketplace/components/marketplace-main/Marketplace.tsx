import Button from "@/modules/core/components/ui/button/Button";
import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import { useNavigate } from "react-router-dom";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";

const getSavedView = (): string => {
  return localStorage.getItem("view") || "grid";
};

const Marketplace = () => {
  const navigate = useNavigate();
  const [view, setView] = useState<string>(getSavedView());

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

  return (
    <>
      <div className="only-section">
        <h2>Plugins Marketplace</h2>
        <MarketplaceFilters
          viewFilter={view as ViewFilter}
          onChange={changeView}
        />
        <section className="cards">
          {[1, 2, 3, 4, 5].map((_, index) => (
            <div className={view === "list" ? "list-card" : ""} key={index}>
              <PluginCard
                pluginType="dca" // todo remove hardcoding once we have the marketplace
                uiStyle={view as ViewFilter}
                id={index.toString()}
                title="DCA Plugin"
                description="The DCA Plugin allows you to dollar cost average into any supported token like Bitcoin. "
              />
            </div>
          ))}
        </section>

        <Button
          size="small"
          type="button"
          styleType="primary"
          onClick={() => navigate(`/plugin-detail/1`)}
        >
          Open Detail view
        </Button>
      </div>
    </>
  );
};

export default Marketplace;
