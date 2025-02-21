import { PluginPolicy, Policy } from "@/modules/policy-form/models/policy";
import { supportedTokens } from "@/modules/dca-plugin/data/tokens";
import { ColumnDef } from "@tanstack/react-table";
import TokenPair from "../shared/token-pair/TokenPair";
import PolicyActions from "./components/PolicyActions";

export const mapData = (
  pluginPolicy: PluginPolicy
): { [key: string]: unknown } => {
  return {
    policyId: pluginPolicy.id,
    pair: [
      pluginPolicy.policy.source_token_id as string,
      pluginPolicy.policy.destination_token_id as string,
    ],
    sell: `${pluginPolicy.policy.total_amount as string} ${supportedTokens[pluginPolicy.policy.source_token_id as string].name}`,
    orders: pluginPolicy.policy.total_orders as string,
    toBuy:
      supportedTokens[pluginPolicy.policy.destination_token_id as string].name,
    orderInterval: `${(pluginPolicy.policy.schedule as Policy).interval} ${(pluginPolicy.policy.schedule as Policy).frequency}`,
    status: true, // todo remove hardcoding when we have this in the DB
  };
};

export const dcaPolicyColumns: ColumnDef<string, any>[] = [
  {
    accessorKey: "pair",
    header: "Pair",
    cell: ({ getValue }) => <TokenPair pair={getValue()} />,
  },
  {
    accessorKey: "sell",
    header: "Sell total",
  },
  {
    accessorKey: "orders",
    header: "Total orders",
  },
  {
    accessorKey: "toBuy",
    header: "To buy",
  },
  {
    accessorKey: "orderInterval",
    header: "Order interval",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ getValue }) => (getValue() ? "Active" : "Inactive"),
  },
  {
    header: "Actions",
    cell: (info: any) => {
      const policyId = info.row.original.policyId;
      return <PolicyActions policyId={policyId} />;
    },
  },
];
