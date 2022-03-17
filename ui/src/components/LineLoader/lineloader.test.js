import React from 'react';
import { mount } from '@cypress/react';
import LineLoader from 'components/LineLoader';

describe("LineLoader", () => {
    it("renders the loader", () => {
        mount(<LineLoader className="test-loader" done={20} total={100} title="test items" displayItems displayPercent />);

        cy.get(".line-loader-container").should("exist");
    });
});