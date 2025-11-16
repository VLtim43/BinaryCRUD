import { h } from "preact";
import "./DataTable.scss";

export interface TableColumn {
  key: string;
  header: string;
  align?: "left" | "center" | "right";
  minWidth?: string;
  render?: (value: any, row: any) => h.JSX.Element | string | number;
}

interface DataTableProps {
  columns: TableColumn[];
  data: any[];
  maxHeight?: string;
  minWidth?: string;
}

export const DataTable = ({ columns, data, maxHeight = "220px", minWidth = "400px" }: DataTableProps) => {
  return (
    <div className="data-table-wrapper" style={{ maxHeight }}>
      <table className="data-table" style={{ minWidth }}>
        <thead>
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                className={`data-table-header data-table-header-${col.align || "left"}`}
                style={{ minWidth: col.minWidth }}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, idx) => (
            <tr key={idx} className={`data-table-row ${idx % 2 === 0 ? "even" : "odd"}`}>
              {columns.map((col) => (
                <td
                  key={col.key}
                  className={`data-table-cell data-table-cell-${col.align || "left"}`}
                >
                  {col.render ? col.render(row[col.key], row) : row[col.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
