// Dependencies
const puppeteer = require('puppeteer');
const { parse, } = require('json2csv');

// Global vars for csv parser
const fields = ['title', 'content'];
const opts = { fields, };

/**
 * Extract review page url
 * @returns {Promise<Object | Error>} - The object containing the review page urls and the total review count
 */
const extractAllReviewPageUrls = async (hotelUrl) => {
    try {

        // Launch the browser
        const browser = await puppeteer.launch({
            headless: true,
            devtools: false,
            defaultViewport: {
                width: 1920,
                height: 1080,
            },
            args: [
                '--disable-gpu',
                '--disable-dev-shm-usage',
                '--disable-setuid-sandbox',
                '--no-sandbox'
            ],
        });

        // Open a new page
        const page = await browser.newPage();

        // Navigate to the page below
        await page.goto(hotelUrl);

        await page.waitForTimeout(5000);

        // Determin current URL
        const currentURL = page.url();

        console.log(`Gathering Info: ${currentURL}`);

        // In browser code
        const getReviewPageUrls = await page.evaluate(() => {

            // All review count
            const totalReviewCount = parseInt(document
                .querySelector('[for=LanguageFilter_1]')
                .innerText.split('(')[1]
                .split(')')[0]
                .replace(',', ''));


            // Default review page count
            let noReviewPages = totalReviewCount / 10;

            // Calculate the last review page
            if (totalReviewCount % 10 !== 0) {
                noReviewPages = ((totalReviewCount - totalReviewCount % 10) / 10) + 1;
            }

            // Get the url of the 2nd page of review. The 1st page is the input link
            let url = false;

            // If there is more than 1 review page
            if (document.getElementsByClassName('pageNum').length > 0) {
                url = document.getElementsByClassName('pageNum')[1].href;
            }

            return { noReviewPages, url, totalReviewCount, };

        });

        // Destructure function outputs
        let { noReviewPages, url, totalReviewCount, } = getReviewPageUrls;

        // Array to hold all the review urls
        const reviewPageUrls = [];

        // If there is more than 1 review page, create the review page url base on the rule below
        if (url) {
            let counter = 0;
            // Replace the url page count till the last page
            while (counter < noReviewPages - 1) {
                counter++;
                url = url.replace(/-or[0-9]*/g, `-or${counter * 10}`);
                reviewPageUrls.push(url);
            }
        }

        // Add the first page url
        reviewPageUrls.unshift(hotelUrl);

        // Information for logging
        const data = {
            count: totalReviewCount,
            pageCount: reviewPageUrls.length,
            urls: reviewPageUrls,
        };
        // console.log(data);

        // Close the browser
        await browser.close();

        return data;

    } catch (err) {
        throw err;
    }
};

/**
 * Scrape the page
 * @param {Array<String>} reviewPageUrls - The review page urls
 * @returns {Promise<Object| Error>} - THe final data
 */
const scrape = async (reviewPageUrls) => {
    try {

        // Launch the browser
        const browser = await puppeteer.launch({
            headless: true,
            devtools: false,
            defaultViewport: {
                width: 1920,
                height: 1080,
            },
            args: [
                '--disable-gpu',
                '--disable-dev-shm-usage',
                '--disable-setuid-sandbox',
                '--no-sandbox'
            ],
        });

        // Open a new page
        const page = await browser.newPage();

        // Array to hold the review info
        const allReviews = [];

        for (let index = 0; index < reviewPageUrls.length; index++) {

            // Navigate to the page below
            await page.goto(reviewPageUrls[index], { waitUntil: 'networkidle2', });

            // Wait for the content to load
            await page.waitForSelector('body');

            // Determin current URL
            const currentURL = page.url();

            console.log(`Scraping: ${currentURL} | ${reviewPageUrls.length - 1 - index} Pages Left`);

            // In browser code
            // Extract comments title
            const commentTitle = await page.evaluate(async () => {

                // Extract a tags
                const commentTitleBlocks = document.getElementsByClassName('fCitC');

                // Array to store the comment titles
                const titles = [];

                // Higher order functions don't work in the browser
                for (let index = 0; index < commentTitleBlocks.length; index++) {
                    titles.push(commentTitleBlocks[index].children[0].innerText);
                }

                return titles;
            });

            // Extract comments text
            const commentContent = await page.evaluate(async () => {

                const commentContentBlocks = document.getElementsByTagName('q');

                // Array use to store the comments
                const comments = [];

                for (let index = 0; index < commentContentBlocks.length; index++) {
                    comments.push(commentContentBlocks[index].children[0].innerText);
                }

                return comments;
            });

            // Format (for CSV processing) the reviews so each review of each page is in an object
            const formatted = commentContent.map((comment, index) => {
                return {
                    title: commentTitle[index],
                    content: comment,
                };
            });

            // Push the formmated review to the  array
            allReviews.push(formatted);

        }

        // Close the browser
        await browser.close();

        // Convert 2D array to 1D
        return allReviews.flat();

    } catch (err) {
        throw err;
    }
};

/**
 * Scraper init function
 * @param {String} hotelUrl - The hotel url 
 * @returns {Promise<Stinrg | Error>} - The CSV 
 */
const start = async (hotelUrl) => {
    try {
        // Extract review page urls
        const { urls, count, } = await extractAllReviewPageUrls(hotelUrl);

        // Scrape the review page
        const results = await scrape(urls);

        // Convert JSON to CSV
        const csv = parse(results, opts);

        return csv;

    } catch (err) {
        throw err;
    }
};

module.exports = start;
