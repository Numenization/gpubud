import json
import requests
from bs4 import BeautifulSoup

def scrape():
    req = requests.get('https://www.microcenter.com/category/4294966937/graphics-cards?storeid=055&rpp=96')
    soup = BeautifulSoup(req.content, 'html.parser')
    productGrid = soup.find('article', {'id': 'productGrid'}).find('ul')
    items = productGrid.find_all('li')

    data = {}
    for item in items:
        link = item.find_next('a')
        data[link['data-id']] = {
            'brand': link['data-brand'],
            'name': link['data-name'],
            'price': link['data-price']
        }

    return data

def main():
    data = scrape()
    for k, v in data.items():
        print(f'{v["name"]}: ${v["price"]}')

if __name__ == "__main__":
    main()