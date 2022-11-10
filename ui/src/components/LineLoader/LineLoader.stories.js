import React from 'react';
import LineLoader from 'components/LineLoader';

export default {
    title: 'LineLoader',
    component: LineLoader,
    argTypes: {
        displayItems: {
            control: {type: "boolean"}
        },
        displayPercent: {
            control: {type: "boolean"}
        },
        title: {
            control: {type: "text"}
        },
        done: {
            control: {type: "number"}
        },
        total: {
            control: {type: "number"}
        },
        className: {
            control: {type: "text"}
        }
    },
    args: {
        displayItems: true,
        displayPercent: true,
        title: "items",
        done: 15,
        total: 60
    }
};

const Template = (args) => <LineLoader {...args} />

export const Default = Template.bind({});