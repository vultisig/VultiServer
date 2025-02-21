import TokenPair from "@/modules/shared/token-pair/TokenPair";
import { TitleFieldProps } from "@rjsf/utils";

export function TitleFieldTemplate(props: TitleFieldProps) {
  const { id, title, uiSchema } = props;

  let source_token_id, destination_token_id;
  if (
    uiSchema?.["ui:options"] &&
    uiSchema["ui:options"]?.["source_token_id"] &&
    uiSchema["ui:options"]?.["destination_token_id"]
  ) {
    source_token_id = uiSchema["ui:options"]["source_token_id"] as string;
    destination_token_id = uiSchema["ui:options"][
      "destination_token_id"
    ] as string;
  }

  return (
    <header style={{ fontSize: "2.125rem" }} id={id}>
      {source_token_id && destination_token_id && (
        <TokenPair pair={[source_token_id, destination_token_id]} />
      )}

      {(!source_token_id || !destination_token_id) && title}
    </header>
  );
}
