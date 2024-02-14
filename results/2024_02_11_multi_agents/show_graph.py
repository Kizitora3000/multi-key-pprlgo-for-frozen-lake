import pandas as pd
import matplotlib.pyplot as plt

# Load the CSV files for each agent
df_agent_1 = pd.read_csv('success_rate_agent_3x3_agent_1.csv')
df_agent_2 = pd.read_csv('success_rate_agent_3x3_agent_2.csv')
df_agent_3 = pd.read_csv('success_rate_agent_3x3_agent_3.csv')

# Add an identifier column for each agent
df_agent_1['Agent'] = 'Agent 1'
df_agent_2['Agent'] = 'Agent 2'
df_agent_3['Agent'] = 'Agent 3'

# Combine all dataframes into a single one
df_combined = pd.concat([df_agent_1, df_agent_2, df_agent_3])

# Pivot the combined dataframe for easier plotting
# This step averages success rates for each episode across different agents if needed
df_pivoted_agents = df_combined.pivot_table(index='Episode', columns='Agent', values='Success Rate', aggfunc='mean').reset_index()

# Setting specific colors for each agent
colors = {'Agent 1': 'black', 'Agent 2': 'blue', 'Agent 3': 'red'}

# Plotting
plt.figure(figsize=(14, 10))
for column in df_pivoted_agents.columns[1:]:  # Skip the 'Episode' column
    plt.plot(df_pivoted_agents['Episode'], df_pivoted_agents[column], label=column, marker='o', color=colors[column])

plt.title('Success Rate in 3x3')
plt.xlabel('Episode')
plt.ylabel('Success Rate')
plt.legend(title='Agent')
plt.grid(True)

plt.show()