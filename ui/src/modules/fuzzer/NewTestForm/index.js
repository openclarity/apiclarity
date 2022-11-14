import React, { useEffect } from 'react';
import { pickBy, isEmpty } from 'lodash';
import { usePrevious } from 'hooks';
import FormWrapper, { validators, TextField, SelectField, useFormikContext, utils } from 'components/Form';
import Text, { TEXT_TYPES } from 'components/Text';
import { TEST_TYPES, AUTH_SCHEME_TYPES } from '../utils';

import './new-test-forms.scss';

const AUTH_FIELDS = {
    username: "username",
    password: "password",
    key: "key",
    value: "value",
    token: "token"
}

const AUTH_FIELDS_WITH_PREFIX = Object.values(AUTH_FIELDS).reduce((acc, curr) => {
    acc = {...acc, [curr]: `auth.${curr}`};
    return acc;
}, {});

const BasicAuthFields = () => (
    <div className="basic-auth-fields">
        <TextField name={AUTH_FIELDS_WITH_PREFIX.username} label="User name" validate={validators.validateRequired} />
        <TextField type="password" name={AUTH_FIELDS_WITH_PREFIX.password} label="Password" validate={validators.validateRequired} />
    </div>
);

const ApiAuthFields = () => (
    <React.Fragment>
        <TextField name={AUTH_FIELDS_WITH_PREFIX.key} label="API key" validate={validators.validateRequired} />
        <TextField name={AUTH_FIELDS_WITH_PREFIX.value} label="API value" validate={validators.validateRequired} />
    </React.Fragment>
);

const BearerAuthFields = () => (
    <TextField type="password" name={AUTH_FIELDS_WITH_PREFIX.token} label="Bearer token" validate={validators.validateRequired} />
);

const AUTH_SCHEME_TO_FIELDS_MAP = {
    [AUTH_SCHEME_TYPES.AuthorizationSchemeBasicAuth.value]: {
        fieldsComponent: BasicAuthFields,
        fields: [AUTH_FIELDS_WITH_PREFIX.username, AUTH_FIELDS_WITH_PREFIX.password]
    },
    [AUTH_SCHEME_TYPES.AuthorizationSchemeApiToken.value]: {
        fieldsComponent: ApiAuthFields,
        fields: [AUTH_FIELDS_WITH_PREFIX.key, AUTH_FIELDS_WITH_PREFIX.value]
    },
    [AUTH_SCHEME_TYPES.AuthorizationSchemeBearerToken.value]: {
        fieldsComponent: BearerAuthFields,
        fields: [AUTH_FIELDS_WITH_PREFIX.token]
    }
};

const FormFields = () => {
    const {values, setFieldValue} = useFormikContext();

    const {authorizationSchemeType} = values.auth;
    const prevAuthorizationSchemeType = usePrevious(authorizationSchemeType);

    useEffect(() => {
        if (!!prevAuthorizationSchemeType && prevAuthorizationSchemeType !== authorizationSchemeType) {
            const {fields=[]} = AUTH_SCHEME_TO_FIELDS_MAP[prevAuthorizationSchemeType] || {};
            fields.forEach(field => setFieldValue(field, ""));
        }
    }, [prevAuthorizationSchemeType, authorizationSchemeType, setFieldValue]);

    const testTypesItems = Object.values(TEST_TYPES);
    const {fieldsComponent: FieldsComponent} = AUTH_SCHEME_TO_FIELDS_MAP[authorizationSchemeType] || {};

    return (
        <React.Fragment>
            <utils.FormNotificationMessage className="new-test-alert" isError>
                Testing, in case of anomalies, could crash your API.
            </utils.FormNotificationMessage>
            <SelectField
                name="depth"
                label="Test type"
                items={testTypesItems}
                tooltipText={(
                    <span>
                        <b>Test types</b>
                        {
                            testTypesItems.map(({value, label, duration}) => <div key={value}>{`${label}: ${duration}  approx.`}</div>)
                        }
                    </span>
                )}
                clearable={false}
                validate={validators.validateRequired}
            />
            <SelectField
                name="auth.authorizationSchemeType"
                label="Authorization scheme"
                items={Object.values(AUTH_SCHEME_TYPES)}
            />
            {!!FieldsComponent && <FieldsComponent />}
        </React.Fragment>
    )
}

const NewTestForm = ({catalogId, onFormSubmitSuccess, onDirtyChanage}) => {
    const initialValues = {
        depth: "",
        auth: {
            // authorizationSchemeType: "",
            type: "",
            ...Object.values(AUTH_FIELDS).reduce((acc, curr) => {
                acc = {...acc, [curr]: ""};
                return acc;
            }, {})
        }
    };
    return (
        <div className="new-fuzzing-test-form">
            <Text type={TEXT_TYPES.TITLE_LARGE} withTopMargin withBottomMargin>New test</Text>
            <Text type={TEXT_TYPES.TABLE_BODY}>
                <div>
                    This will test: APIs providing invalid, unexpected or random data as inputs to a computer program.<br />
                    This APIs are then monitored for exceptions such as crashes, failing built-in code assertions, or potential memory leaks.<br /><br />
                    Once the testing has been completed, the risk findings will be updated accordingly.
                </div>
            </Text>
            <FormWrapper
                initialValues={initialValues}
                submitUrl={`modules/fuzzer/fuzz/${catalogId}/start`}
                getSubmitParams={formValues => {
                    const {auth} = formValues;
                    // TODO: where is authorizationSchemeType being set? Need to remove it.
                    auth.type = auth.authorizationSchemeType;
                    const cleanAuth = pickBy(auth, value => value !== "");

                    return {
                        submitData: {...formValues, auth: isEmpty(cleanAuth) ? null : cleanAuth}
                    }
                }}
                onSubmitSuccess={onFormSubmitSuccess}
                onDirtyChanage={onDirtyChanage}
            >
                <FormFields />
            </FormWrapper>
        </div>
    )
}

export default NewTestForm;
