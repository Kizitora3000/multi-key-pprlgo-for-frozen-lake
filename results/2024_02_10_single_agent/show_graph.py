import matplotlib.pyplot as plt
import pandas as pd

# Load the CSV files
df_4x4 = pd.read_csv('success_rate_4x4.csv')
df_5x5 = pd.read_csv('success_rate_5x5.csv')
df_6x6 = pd.read_csv('success_rate_6x6.csv')

# Set the figure size for better visibility
plt.figure(figsize=(14, 8))

# Plot each DataFrame
plt.plot(df_4x4['Episode'], df_4x4['Success Rate'], label='4x4', color='blue')
plt.plot(df_5x5['Episode'], df_5x5['Success Rate'], label='5x5', color='red')
plt.plot(df_6x6['Episode'], df_6x6['Success Rate'], label='6x6', color='green')

# Setting the labels and title
plt.xlabel('Episode')
plt.ylabel('Success Rate')
plt.title('Success Rate by Puzzle Size Over Episodes')
plt.legend()

# Show the plot
plt.grid(True)
plt.show()
