import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.scale as mscale
import matplotlib.transforms as mtransforms
import matplotlib.ticker as ticker
import numpy as np

# 横軸を0からスタートさせたうえで0と1の間隔を狭めるようスケールを設定する
class CustomScale(mscale.ScaleBase):
    name = 'custom'

    def __init__(self, axis, **kwargs):
        mscale.ScaleBase.__init__(self, axis)

    def get_transform(self):
        return self.CustomTransform()

    def set_default_locators_and_formatters(self, axis):
        axis.set_major_locator(ticker.MaxNLocator(integer=True))  # 整数値のみを配置

    class CustomTransform(mtransforms.Transform):
        input_dims = 1
        output_dims = 1
        is_separable = True

        def transform_non_affine(self, a):
            # 0から1の間を圧縮する変換
            return np.where(a <= 1, a * 0.3, 0.3 + (a - 1) * 1.0)

        def inverted(self):
            return CustomScale.InvertedCustomTransform()

    class InvertedCustomTransform(mtransforms.Transform):
        input_dims = 1
        output_dims = 1
        is_separable = True

        def transform_non_affine(self, a):
            return np.where(a <= 0.3, a / 0.3, 1 + (a - 0.3) / 1.0)

        def inverted(self):
            return CustomScale.CustomTransform()

# スケールを登録
mscale.register_scale(CustomScale)

# CSVファイルの読み込み
def load_data(file_path):
    df = pd.read_csv(file_path)
    state_counts = df["状態数"].values
    num_people = df.columns[1:].astype(int)
    
    def time_to_seconds(time_str):
        if 'm' in time_str:
            minutes, seconds = time_str.split('m')
            return int(minutes) * 60 + float(seconds.replace('s', ''))
        else:
            return float(time_str.replace('s', ''))
    
    times = {state_counts[i]: [time_to_seconds(df.iloc[i, j]) for j in range(1, len(df.columns))] for i in range(len(state_counts))}
    return num_people, times

# グラフの作成
def plot_processing_time(file_path):
    num_people, times = load_data(file_path)
    
    # plt.rcParams['font.family'] = 'MS Gothic'
    plt.rcParams['font.size'] = 28

    # カラーコードの設定
    colors = {
        16: 'red',
        25: 'blue',
        36: 'green'
    }

    plt.figure(figsize=(8, 6))
    for state, time_list in times.items():
        plt.plot(num_people, time_list, marker='o', 
                color=colors[state],  # カラーコードを指定
                label=f'$N_s$: {state}')
    
    plt.xlabel("Number of users")
    plt.ylabel("Processing time [seconds]")
    plt.legend(loc='lower right')
    plt.grid()
    
    # カスタムスケールを設定
    plt.gca().set_xscale('custom')
    
    # 整数値のみの目盛りを設定
    plt.gca().yaxis.set_major_locator(ticker.MaxNLocator(integer=True))
    
    # 軸の範囲を0から設定
    plt.xlim(0, max(num_people) * 1.1)
    plt.ylim(0, max(max(time_list) for time_list in times.values()) * 1.1)
    
    plt.show()

file_path = "performance.csv"
plot_processing_time(file_path)