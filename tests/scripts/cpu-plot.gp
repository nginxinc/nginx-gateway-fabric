set terminal png size 800,600
set title "CPU Usage"
set datafile separator ","
set output outputfile . ""

# X-axis settings
set xlabel "Timestamp"
set xdata time
set timefmt "%s"
set format x "%M:%S"
set xrange [*:*]
set xtics nomirror

# Y-axis settings
set yrange [0:*]
set ylabel "CPU Usage (core seconds)"
set format y "%.2f"

# Plotting data
plot inputfile using 1:2 with lines lw 2 notitle
