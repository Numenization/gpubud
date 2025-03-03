import requests
import json
from bs4 import BeautifulSoup

def scrape():
    req = requests.get('https://www.microcenter.com/category/4294966937/graphics-cards?storeid=055&rpp=96')
    soup = BeautifulSoup(req.content, 'html.parser')
    productGrid = soup.find('article', {'id': 'productGrid'}).find('ul')
    items = productGrid.find_all('li')

    data = {}
    for item in items:
        link = item.find_next('a')
        link2 = link.find_next('a')

        parsed_name = parse_name(link['data-name'])

        if parsed_name is None:
            continue

        data[link['data-id']] = {
            'manufacturer': link['data-brand'],
            'name': link['data-name'],
            'price': link['data-price'],
            'brand': parsed_name['brand'],
            'line': parsed_name['line'],
            'model': parsed_name['model'],
            'link': f'https://www.microcenter.com{link2["href"]}'
        }

    return data

def parse_name(name):
    tokens = name.split(' ')
    data = {}

    # get brand
    data['brand'] = tokens[0]

    # check for GeForce RTX cards
    if tokens[1] == 'GeForce' and tokens[2] == 'RTX':
        data['line'] = 'GeForce RTX'
    # check for Radeon RX cards
    elif tokens[1] == 'Radeon' and tokens[2] == 'RX':
        data['line'] = 'Radeon RX'
    else:
        # we don't want any cards that arent RTX or RX
        return None

    # get model of card
    data['model'] = tokens[3]

    return data


def main():
    data = scrape()
    print(json.dumps(data, indent=2, sort_keys=True))

if __name__ == "__main__":
    main()