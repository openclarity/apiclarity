import React, { useCallback, useEffect, useRef } from 'react';
import { isEmpty } from 'lodash';
import Checkbox from 'components/Checkbox';

import './columns-select-panel.scss';

const ColumnsList = ({columns}) => (
    <React.Fragment>
        {
            columns.map(column => {
                const {checked, onChange} = column.getToggleHiddenProps();
                const {id, alwaysShow} = column;

                return (
                    <Checkbox
                        key={id}
                        name={id}
                        checked={checked}
                        value={id}
                        title={column.render('Header')}
                        onChange={onChange}
                        disabled={alwaysShow}
                        small
                    />
                );
            })
        }
    </React.Fragment>
);

const ColumnsListWithHeaders = ({headerColumnNames, columns}) => (
    <React.Fragment>
        {
            headerColumnNames.map(headerName => (
                <div key={headerName} className="header-columns-wrapper">
                    <div className="header-column-title">{headerName}</div>
                    <div className="header-columns">
                        <ColumnsList columns={columns.filter(column => column.parent.Header === headerName)} />
                    </div>
                </div>
            ))
        }
    </React.Fragment>
);

const ColumnsSelectPanel = ({columns, headerColumnNames, onClose, columnsIconClassName}) => {
    const columnsPanelRef = useRef();

    const handleClick = useCallback(({target}) => {
        if (target.parentElement.classList.contains(columnsIconClassName)) {
            //clicked the columns icon
            return;
        }

        if (columnsPanelRef.current.contains(target)) {
            //clicked inside the panel
            return;
        }

        onClose();
    }, [onClose, columnsIconClassName]);

    useEffect(() => {
        document.addEventListener("mousedown", handleClick);

        return () => {
            document.removeEventListener("mousedown", handleClick);
        };
      }, [handleClick]);

    return (
        <div className="columns-select-panel-container" ref={columnsPanelRef}>
            {
                isEmpty(headerColumnNames) ? <ColumnsList columns={columns} /> :
                    <ColumnsListWithHeaders columns={columns} headerColumnNames={headerColumnNames} />
            }
        </div>
    );
};

export default ColumnsSelectPanel;
