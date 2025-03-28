import Accordion from "@/modules/core/components/ui/accordion/Accordion";
import "./Summary.css";

type Row = { key: string; value: string };

type SummaryProps = {
  title: string;
  data: Row[];
};
const Summary = ({ title = "Summary", data }: SummaryProps) => {
  return (
    <Accordion header={null} expandButton={{ text: title }}>
      <div className="summary-content">
        {data.map(({ key, value }) => (
          <div key={key} className="summary-row">
            <span>{key}</span>&nbsp;<span>{value}</span>
          </div>
        ))}
      </div>
    </Accordion>
  );
};

export default Summary;
