import { PluginPolicy, Policy } from "@/modules/policy/models/policy";
import { supportedTokens } from "@/modules/shared/data/tokens";
import { ColumnDef } from "@tanstack/react-table";
import TokenPair from "../../shared/token-pair/TokenPair";
import PolicyActions from "../components/policy-actions/PolicyActions";

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
    status: pluginPolicy.active,
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
    cell: ({ getValue }) =>
      getValue() ? (
        <span
          style={{
            display: "flex",
            alignItems: "center",
            whiteSpace: "nowrap",
            color: "#13C89D",
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              backgroundColor: "#13C89D",
              borderRadius: "50%",
            }}
          ></span>
          &nbsp; Active
        </span>
      ) : (
        <span
          style={{
            display: "flex",
            alignItems: "center",
            whiteSpace: "nowrap",
            color: "#8295AE",
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              backgroundColor: "#8295AE",
              borderRadius: "50%",
            }}
          ></span>
          &nbsp; Inactive
        </span>
      ),
  },
  {
    header: "Actions",
    cell: (info: any) => {
      const policyId = info.row.original.policyId;
      return <PolicyActions policyId={policyId} />;
    },
  },
];
