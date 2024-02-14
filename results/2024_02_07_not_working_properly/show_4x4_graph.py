import pandas as pd
import matplotlib.pyplot as plt

# Load the CSV file
data = pd.read_csv('success_rate4x4.csv')  # Replace 'path_to_your_file.csv' with the actual file path

# Filter the data to include only up to 5000 trials
filtered_data = data[data['Trials'] <= 5000]

# Plotting
plt.figure(figsize=(10, 6))
plt.plot(filtered_data['Trials'], filtered_data['Goal Rate'], color='black')
plt.title('4×4', fontsize=16)
plt.xlabel('Episode', fontsize=14)
plt.ylabel('Success rate', fontsize=14)
plt.grid(True)
plt.legend().set_visible(False)  # Hiding the legend as requested

# Display the plot
plt.show()