import Button from "@/modules/core/components/ui/button/Button";
import { useNavigate } from "react-router-dom";
import logo from "../../../../assets/DCA-image.png"; // Adjust path based on file location
import "./PluginCard.css";
import { ViewFilter } from "@/modules/marketplace/models/marketplace";

type PluginCardProps = {
  id: string;
  title: string;
  description: string;
  uiStyle: ViewFilter;
};

const PluginCard = ({ uiStyle, id, title, description }: PluginCardProps) => {
  const navigate = useNavigate();

  return (
    <div className={`plugin ${uiStyle}`}>
      <div className={uiStyle === "grid" ? "" : "info-group"}>
        <img src={logo} alt={title} />

        <div className="plugin-info">
          <h3>{title}</h3>
          <p>{description}</p>
        </div>
      </div>

      <Button
        style={uiStyle === "grid" ? { width: "100%" } : {}}
        size={uiStyle === "grid" ? "small" : "mini"}
        type="button"
        styleType="primary"
        onClick={() => navigate(`/plugin-policy/${id}`)}
      >
        See details
      </Button>
    </div>
  );
};

export default PluginCard;
