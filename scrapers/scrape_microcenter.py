import requests
import json
import datetime
import argparse
from bs4 import BeautifulSoup

def scrape(source):
    req = requests.get(source)
    soup = BeautifulSoup(req.content, 'html.parser')
    productGrid = soup.find('article', {'id': 'productGrid'}).find('ul')
    items = productGrid.find_all('li')

    data = {
        'timestamp': datetime.datetime.now().strftime('%m-%d-%Y %H:%M:%S'),
        'source': source,
        'gpus': []
    }
    for item in items:
        left = item.find('div', {'class': 'result_left'})
        right = item.find('div', {'class': 'result_right'})

        link = left.find_next('a')
        link2 = link.find_next('a')
        count_span = right.find('span', {'class': 'inventoryCnt'})

        stock = 'n/a'
        if count_span is not None:
            stock = count_span.contents[0].strip()

        parsed_name = parse_name(link['data-name'])

        data['gpus'].append({
            'manufacturer': link['data-brand'],
            'name': link['data-name'],
            'price': link['data-price'],
            'id': link['data-id'],
            'brand': parsed_name['brand'],
            'line': parsed_name['line'],
            'model': parsed_name['model'],
            'stock': stock,
            'link': f'https://www.microcenter.com{link2["href"]}'
        })

    return data

def parse_name(name):
    tokens = iter(name.split(' '))
    data = {}

    cur = next(tokens)

    # get brand
    data['brand'] = cur

    # build line name
    cur = next(tokens)
    line = ''
    while cur.isalpha() == True:
        line = f'{line} {cur}'
        cur = next(tokens)

    # build model name
    model = cur
    cur = next(tokens)
    if cur == 'XT' or cur == 'Ti':
        model = f'{model} {cur}'
        cur = next(tokens)

    data['line'] = line
    data['model'] = model

    return data


def main():
    parser = argparse.ArgumentParser(description='Scrape Microcenter website for GPU listings')
    parser.add_argument('-s', '--source', required=True, dest='source', action='store', help='Source URL to scrape')
    parser.add_argument('-p', '--pretty', dest='pretty', action='store_true', help='Format the result JSON for human readability')
    args = parser.parse_args()

    data = scrape(args.source)

    if args.pretty:
        print(json.dumps(data, indent=2, sort_keys=True))
    else:
        print(json.dumps(data))

if __name__ == "__main__":
    main()