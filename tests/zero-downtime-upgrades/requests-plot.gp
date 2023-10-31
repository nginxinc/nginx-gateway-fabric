set terminal png size 800,600
set output "graph.png"
set title "Request Outcomes Over Time"

set xdata time
set timefmt "%Y-%m-%d %H:%M:%S"
set datafile separator ","

# Y-axis settings
set yrange [-0.5:1.5]   # Provide some padding around 0 and 1 for better visualization
set ytics ("Failed" 0, "Success" 1)
set grid ytics           # Gridlines for Y

# Define the palette: 0 for red (Failure) and 1 for green (Success)
set palette defined (0 "red", 1 "green")

# Hide the colorbox
unset colorbox

# Plotting data
plot "results.csv" using 1:2:2 with points palette pointtype 7 pointsize 1.5 title "Request Status"
