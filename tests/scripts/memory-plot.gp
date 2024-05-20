# Define a function to convert bytes to Mebibytes
bytes_to_MiB(bytes) = bytes / (1024.0 * 1024.0)

set terminal png size 800,600
set title "Memory Usage"
set datafile separator ","
set output outputfile . ""

# X-axis settings
set xlabel "Timestamp"
set xdata time
set timefmt "%s"
set format x "%M:%S"
set xrange [*:*]  # Specify a range covering all timestamps

# Y-axis settings
set yrange [0:*]
set ylabel "Memory Usage (MiB)"

# Plotting data
plot inputfile using 1:(bytes_to_MiB($2)) with lines lw 2 notitle
