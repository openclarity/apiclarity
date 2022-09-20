BEGIN {
    # Set default value if unspecified
    if (APICLARITY_PLUGINS_API_VERSION ~ "^$") {
        APICLARITY_PLUGINS_API_VERSION="v0.0.0-20220915093602-8a11adcdb9e1"
    }
}
{
    if (/output_path/) {
        print "  output_path: .";
    } else if (/exporters:/) {
        print $0, "\n",
            " - gomod: github.com/openclarity/apiclarity/plugins/otel-collector/apiclarityexporter v0.0.0";
    } else {
        print;
    }
}
END {
    print "\nreplaces:\n",
        " - github.com/openclarity/apiclarity/plugins/otel-collector/apiclarityexporter v0.0.0 => ../apiclarityexporter\n",
        " - github.com/openclarity/apiclarity/plugins/api v0.0.0 => github.com/openclarity/apiclarity/plugins/api",APICLARITY_PLUGINS_API_VERSION,"\n";
}