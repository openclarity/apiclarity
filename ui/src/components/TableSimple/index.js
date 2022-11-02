import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';

import './table-simple.scss';

const TableSimple = ({ headers, rows, className, emptyText = "No data", hideBorder = false }) => {
    return (
        <table className={classnames("table-simple", className, { "no-border": hideBorder })}>
            {!isEmpty(headers) &&
                <thead>
                    <tr>
                        {
                            headers.map((header, index) => <th key={index}>{header}</th>)
                        }
                    </tr>
                </thead>
            }
            <tbody>
                {
                    rows.map((rowCells, index) => (
                        <tr key={index}>
                            {
                                rowCells.map((cell, index) => <td key={index}>{cell}</td>)
                            }
                        </tr>
                    ))
                }
                {isEmpty(rows) && <tr><td colSpan={headers.length}>{emptyText}</td></tr>}
            </tbody>
        </table>
    )
}

export default TableSimple;
