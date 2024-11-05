import matplotlib.pyplot as plt
import pandas as pd

# Create a sample dataset
data = {'x': [1, 2, 3, 4, 5], 'y': [1, 4, 9, 16, 25]}
df = pd.DataFrame(data)

# Generate a plot
fig, ax = plt.subplots()
df.plot(x='x', y='y', ax=ax)
ax.set_title('Sample Line Plot')

# Assign the figure to the output variable
output = fig
