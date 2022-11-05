import pandas as pd

airbnb = pd.read_csv('listings.csv')

# Change series type
airbnb['last_review'] = pd.to_datetime(airbnb['last_review'])

# Add new series
airbnb['year'] = airbnb['last_review'].dt.year

# Clean series
airbnb['name'] = airbnb['name'].str.strip()
airbnb['name_lower'] = airbnb['name'].str.lower()

# Calculations
airbnb['price'] = airbnb['price'].replace('[$,]', '', regex=True).astype(float)
airbnb['min_revenue'] = airbnb['minimum_nights'] * airbnb['price']

print(airbnb[['minimum_nights', 'price', 'min_revenue']].head())
print(airbnb['price'].mean())
print(airbnb[['room_type', 'year', 'price']].groupby(['room_type', 'year'], as_index=False).mean())

print(airbnb[airbnb['price'] < 1000].head())
