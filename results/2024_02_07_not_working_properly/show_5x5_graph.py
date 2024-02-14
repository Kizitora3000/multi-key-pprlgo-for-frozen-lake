import pandas as pd
import matplotlib.pyplot as plt

# Load the CSV file
data_5x5 = pd.read_csv('success_rate5x5.csv')  # Replace 'path_to_your_5x5_file.csv' with the actual file path

# Plotting for 5x5 data
plt.figure(figsize=(10, 6))
plt.plot(data_5x5['Trials'], data_5x5['Goal Rate'], color='red')
plt.title('5×5', fontsize=16)
plt.xlabel('Episode', fontsize=14)
plt.ylabel('Success rate', fontsize=14)
plt.grid(True)
plt.legend().set_visible(False)  # Hiding the legend as requested

# Display the plot
plt.show()