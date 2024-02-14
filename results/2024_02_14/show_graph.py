import matplotlib.pyplot as plt
import pandas as pd

# Load your data
df_mkpprl_4x4 = pd.read_csv('MKPPRL_success_rate_4x4.csv')
df_mkpprl_5x5 = pd.read_csv('MKPPRL_success_rate_5x5.csv')
df_mkpprl_6x6 = pd.read_csv('MKPPRL_success_rate_6x6.csv')
df_pprl_4x4 = pd.read_csv('PPRL_success_rate_4x4.csv')
df_pprl_5x5 = pd.read_csv('PPRL_success_rate_5x5.csv')
df_pprl_6x6 = pd.read_csv('PPRL_success_rate_6x6.csv')

# Plotting
plt.figure(figsize=(10, 8))

# MKPPRL data plotting with color specifications
plt.plot(df_mkpprl_4x4['Episode'], df_mkpprl_4x4['Success Rate'], label='MKPPRL 4x4 Grid', linestyle='solid', color='blue', markersize=4)
plt.plot(df_mkpprl_5x5['Episode'], df_mkpprl_5x5['Success Rate'], label='MKPPRL 5x5 Grid', linestyle='solid', color='red', markersize=4)
plt.plot(df_mkpprl_6x6['Episode'], df_mkpprl_6x6['Success Rate'], label='MKPPRL 6x6 Grid', linestyle='solid', color='green', markersize=4)


# PPRL data plotting with color specifications
plt.plot(df_pprl_4x4['Episode'], df_pprl_4x4['Success Rate'], label='PPRL 4x4 Grid',  linestyle='dashed', color='#6196F2', markersize=4)
plt.plot(df_pprl_5x5['Episode'], df_pprl_5x5['Success Rate'], label='PPRL 5x5 Grid',  linestyle='dashed', color='#FA6964', markersize=4)
plt.plot(df_pprl_6x6['Episode'], df_pprl_6x6['Success Rate'], label='PPRL 6x6 Grid', linestyle='dashed', color='#66FF66', markersize=4)

# Adding titles, labels, and legend
plt.xlabel('Episode')
plt.ylabel('Success Rate')
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.show()