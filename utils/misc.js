// Dependencies
const { promises: { access, }, readdirSync, writeFileSync, } = require('fs');
const { parse, } = require('json2csv');

/**
 * Check if the given file exists
 * @param {String} filePath 
 * @returns {Promise<Boolean>}
 */
const fileExists = async (filePath) => {
    try {
        await access(filePath);
        return true;
    } catch (err) {
        return false;
    }
};

/**
 * Combine all JSON files in the data directory into a JSON array of object
 * @returns {Array<Object>}
 */
const combine = () => {
    try {
        const allFiles = readdirSync('../data/');

        const extracted = allFiles.map(file => {
            // eslint-disable-next-line global-require
            const fileContent = require(`../data/${file}`);
            const { restoName, restoId, position, allReviews, } = fileContent;
            return { restoName, restoId, position, allReviews, };
        })
            .sort((a, b) => a.position - b.position)
            .map(resto => {
                const { restoName, restoId, position, allReviews, } = resto;
                return allReviews.map(review => {
                    review.restoName = restoName;
                    review.restoId = restoId;
                    review.position = position;
                    return review;
                });
            })
            .flat()
            .map(review => {
                const { restoName, restoId, rating, dateOfVist, ratingDate, title, content, } = review;

                return { restoName, restoId, title, content, rating, dateOfVist, ratingDate, };
            });

        return extracted;

    } catch (err) {
        throw err;
    }
};

/**
 * Convert JSON input to CSV
 * @param {Array<Object>} jsonInput - The JSON array of review and restaurant objects
 * @returns {String} - The CSV string
 */
const reviewJSONToCsv = (jsonInput) => {
    try {

        const fields = Object.keys(jsonInput[0]);
        const opts = { fields, };

        // Convert JSON to CSV
        const csv = parse(jsonInput, opts);

        // Write the CSV to a file
        writeFileSync('../reviews.csv', csv);

    } catch (err) {
        throw err;
    }
};


module.exports = { fileExists, combine, reviewJSONToCsv, };

console.log(reviewJSONToCsv(combine()));