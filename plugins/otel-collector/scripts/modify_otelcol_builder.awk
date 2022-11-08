#!/usr/bin/awk -f
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
        " - github.com/openclarity/apiclarity/plugins/otel-collector/apiclarityexporter v0.0.0 => ./otel-collector/apiclarityexporter\n",
        " - github.com/openclarity/apiclarity/plugins/api v0.0.0 => ./api\n";
}