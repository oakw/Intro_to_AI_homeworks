# Just a simple script to visualize the path of the agent in the grid environment.
# Usage: go run . optimal input.csv | python visualize.py input.csv
# Generated using ChatGPT
import sys
import numpy as np
import csv
import ast
import matplotlib.pyplot as plt
import matplotlib.colors as mcolors
import random

def read_csv(file_path):
    """ Reads the grid and configuration from a CSV file. """
    with open(file_path, newline='') as csvfile:
        reader = list(csv.reader(csvfile))
        # Read initial conditions
        x_coords = reader[0][0].split("#")
        y_coords = reader[1][0].split("#")
        battery = int(reader[2][0].split("#")[0])
        movement_cost = int(reader[3][0].split("#")[0])
        cleaning_cost = int(reader[4][0].split("#")[0])
        start_x = int(x_coords[0])
        start_y = int(y_coords[0])

        # Read grid (starting from line 6)
        grid = np.array([[int(x) for x in row] for row in reader[5:]])

    return grid, (start_x, start_y), battery, movement_cost, cleaning_cost

def parse_stdin():
    """ Reads the agent's movement path from stdin. """
    path = []

    for line in sys.stdin:
        line = line.strip()
        
        if line.startswith("Moved to"):
            coords = ast.literal_eval(line.split("Moved to ")[1])
            # Reverse coordinates to match the grid
            coords = (coords[1], coords[0])
            path.append(coords)

    return path

def visualize(grid, path, energy, movement_cost, cleaning_cost, name="grid"):
    """ Displays the grid with scores and visualizes the agent's path with a gradient and directional arrows. """
    rows, cols = grid.shape

    # Fixed figure size for consistent visualization
    # fig, ax = plt.subplots(figsize=(8, 8 * (rows / cols) if cols > rows else 8 * (cols / rows)))
    fig, ax = plt.subplots(figsize=(8, 8))  # Fixed window size
    # Display costs above the grid (annotation)
    ax.text(0, rows + 0.2, f"B = {energy}, E = {movement_cost}, V = {cleaning_cost}, {name}", fontsize=12, ha='left', va='center', color='black', weight='bold')

    # Track visit counts
    visit_count = np.zeros_like(grid, dtype=int)
    for r, c in path:
        visit_count[r, c] += 1

    # Draw grid and walls
    for r in range(rows):
        for c in range(cols):
            if grid[r, c] == 9001:
                ax.add_patch(plt.Rectangle((c, rows-r-1), 1, 1, color='black'))  # Walls
            else:
                ax.add_patch(plt.Rectangle((c, rows-r-1), 1, 1, color='lightgray', edgecolor='gray'))
                if grid[r, c] > 0:
                    ax.text(c + 0.5, rows - r - 0.2, str(grid[r, c]), fontsize=10,  # Score on top
                            ha='center', va='center', color='black', weight='bold')

    # Generate gradient colors for the path
    cmap = plt.get_cmap('hsv')  # Choose a color gradient
    norm = mcolors.Normalize(vmin=0, vmax=len(path))  # Normalize steps
    path_colors = [cmap(norm(i)) for i in range(len(path))]

    # Plot the path with arrows
    for i in range(1, len(path)):
        x1, y1 = path[i-1][1] + 0.5, rows - path[i-1][0] - 0.5
        x2, y2 = path[i][1] + 0.5, rows - path[i][0] - 0.5
        ax.plot([x1, x2], [y1, y2], marker='o', color=path_colors[i], markersize=6, linewidth=2)

        # Random position along the line for the arrow
        t = random.uniform(0.3, 0.7)  # Random position between 30% and 70% of the segment
        arrow_x = x1 + t * (x2 - x1)
        arrow_y = y1 + t * (y2 - y1)

        # Add arrow indicating direction
        ax.arrow(arrow_x, arrow_y, (x2 - x1) * 0.2, (y2 - y1) * 0.2, head_width=0.2, 
                 head_length=0.2, fc=path_colors[i], ec=path_colors[i])

    # Mark visit counts
    for r in range(rows):
        for c in range(cols):
            if visit_count[r, c] > 1:
                ax.text(c + 0.5, rows - r - 0.8, str(visit_count[r, c]), fontsize=10,
                        ha='center', va='center', color='blue', weight='bold')

    # Mark start and end
    ax.text(path[0][1] + 0.5, rows - path[0][0] - 0.5, 'S', fontsize=12,
            ha='center', va='center', color='white', weight='bold')
    ax.text(path[-1][1] + 0.5, rows - path[-1][0] - 0.5, 'E', fontsize=12,
            ha='center', va='center', color='yellow', weight='bold')

    # Adjust limits and hide axes
    ax.set_xlim(0, cols)
    ax.set_ylim(0, rows)
    ax.set_xticks([])
    ax.set_yticks([])
    ax.set_frame_on(False)
    ax.set_aspect('equal')  # Maintain square cells

    # plt.show()
    plt.savefig(name.replace('/', '_').replace('\\', '_').strip('._') + '.png', dpi=300)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python script.py <csv_file>")
        sys.exit(1)

    csv_file = sys.argv[1]
    grid, start_pos, battery, movement_cost, cleaning_cost = read_csv(csv_file)
    path = parse_stdin()

    visualize(grid, path, battery, movement_cost, cleaning_cost, name=csv_file)
