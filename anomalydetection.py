# python outlier detection

import warnings
import numpy as np
import pandas as pd
from pyod.models.mad import MAD
from pyod.models.knn import KNN
from pyod.models.lof import LOF
import matplotlib.pyplot as plt
from sklearn.ensemble import IsolationForest

# sample data for anomaly detection
data_values = [['2021-05-1', 300.0],
       ['2021-05-2', 305.0],
       ['2021-05-3', 500.0],
       ['2021-05-4', 600.0],
       ['2021-05-5', 300.0],
       ['2021-05-6', 300.0],
       ['2021-05-7', 350.0],
       ['2021-05-8', 600.0],
       ['2021-05-9', 450.0],
       ['2021-05-10', 200.0],
       ['2021-05-11', 136.0],
       ['2021-05-12', 126.0],
       ['2021-05-13', 347.0],
       ['2021-05-14', 462.66],
       ['2021-05-15', 543.0],
       ['2021-05-16', 462.66],
       ['2021-05-17', 207.0],
       ['2021-05-18', 502.0],
       ['2021-05-19', 589.0],
       ['2021-05-20', 404.0],
       ['2021-05-21', 593.0],
       ['2021-05-22', 267.0],
       ['2021-05-23', 880.0],
       ['2021-05-24', 361.0],
       ['2021-05-25', 404.0],
       ['2021-05-26', 209.0],
       ['2021-05-27', 461.0],
       ['2021-05-28', 866.0],
       ['2021-05-29', 886.0],
       ['2021-05-30', 336.0],
       ['2021-05-31', 251.0],
       ['2021-06-1', 270.0],
       ['2021-06-2', 620.0],
       ['2021-06-3', 317.0]]
       
data = pd.DataFrame(data_values , columns=['date', 'amount'])

def fit_model(model, data, column='amount'):
    # fit the model and predict it
    df = data.copy()
    data_to_predict = data[column].to_numpy().reshape(-1, 1)
    predictions = model.fit_predict(data_to_predict)
    df['Predictions'] = predictions
    
    return df

def plot_anomalies(df, x='date', y='amount'):

    # categories will be having values from 0 to n
    # for each values in 0 to n it is mapped in colormap
    categories = df['Predictions'].to_numpy()
    colormap = np.array(['g', 'r'])

    f = plt.figure(figsize=(12, 4))
    f = plt.scatter(df[x], df[y], c=colormap[categories])
    f = plt.xlabel(x)
    f = plt.ylabel(y)
    f = plt.xticks(rotation=90)
    #plt.show()
    plt.savefig("figure1")


def find_anomalies(value, lower_threshold, upper_threshold):
    
    if value < lower_threshold or value > upper_threshold:
        return 1
    else: return 0

def iqr_anomaly_detector(data, column='amount', threshold=1.1):
    
    df = data.copy()
    quartiles = dict(data[column].quantile([.25, .50, .75]))
    quartile_3, quartile_1 = quartiles[0.75], quartiles[0.25]
    iqr = quartile_3 - quartile_1

    lower_threshold = quartile_1 - (threshold * iqr)
    upper_threshold = quartile_3 + (threshold * iqr)

    print(f"Lower threshold: {lower_threshold}, \nUpper threshold: {upper_threshold}\n")
    
    df['Predictions'] = data[column].apply(find_anomalies, args=(lower_threshold, upper_threshold))
    return df
  

"""KNN Based Outlier Detection"""
knn_model = KNN()
knn_df = fit_model(knn_model, data)
plot_anomalies(knn_df)