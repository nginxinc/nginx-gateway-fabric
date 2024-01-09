set terminal png size 800,600
set title "Request Outcomes Over Time"
set datafile separator ","
set output outputfile . ""

# X-axis settings
set xdata time
set format "%M:%S"

# Y-axis settings
set yrange [-50:550]
set ytics ("0" 0, "200" 200, "4xx" 403 404, "5xx" 500 502)
set grid ytics

# Plotting data (vegeta timestamp is in nanoseconds, so convert to seconds)
plot inputfile using ($1/1000000000):2:2 pointtype 7 pointsize 1.5 notitle
