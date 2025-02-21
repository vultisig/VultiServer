import {
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { usePolicies } from "@/modules/policy-form/context/PolicyProvider";
import { mapData, dcaPolicyColumns } from "../dcaTableSchema"; // todo these should be dynamic once we have the marketplace
import PolicyFilters from "./PolicyFilters";

const columns = [...dcaPolicyColumns];

const PolicyTable = () => {
  const [data, setData] = useState<any>(() => []);
  const { policyMap } = usePolicies();

  useEffect(() => {
    const policies = [];
    for (const [_, value] of policyMap) {
      policies.push(mapData(value));
    }
    setData(policies);
  }, [policyMap]);

  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]); // can set initial column filter state here

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
    },
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(), // needed for client-side filtering
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div>
      <PolicyFilters onFiltersChange={setColumnFilters} />
      <table bgcolor="#061B3A" cellPadding={7}>
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th key={header.id}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
          {table.getRowModel().rows.length === 0 && (
            <tr>
              <td>Nothing to see here yet.</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default PolicyTable;
