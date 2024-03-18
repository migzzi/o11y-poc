import "./tracing";

import Axios from "axios";
import Express, { NextFunction, Request, Response } from "express";
import Morgan from "morgan";

const axios = Axios.create();
export const app = Express();
const port = process.env.APP_PORT || 3000;

function asyncHandler(fn: Function) {
  return (req: Request, res: Response, next: NextFunction) => {
    return Promise.resolve(fn(req, res, next)).catch(next);
  };
}

function failWithProbabilityMW(
  probability: number
): (req: Request, res: Response, next: NextFunction) => void {
  return (req: Request, res: Response, next: NextFunction) => {
    if (Math.random() < probability) {
      return res.status(500).send("Internal Server Error");
    }

    return next();
  };
}

app.use(Express.json());
app.use(Express.urlencoded({ extended: true }));
app.use(Morgan("short"));

app.get("/health", (req, res) => {
  res.send("OK");
});

app.get(
  "/home",
  asyncHandler(async (req: Request, res: Response) => {
    const [productsResponse, categoriesResponse] = await Promise.all([
      axios.get(`${process.env.PRODUCTS_SERVICE_URL}/products`),
      axios.get(`${process.env.CATEGORIES_SERVICE_URL}/categories`),
    ]);

    return res.json({
      products: productsResponse.data,
      categories: categoriesResponse.data,
    });
  })
);

app.get("/might-fail", failWithProbabilityMW(0.1), async (req, res) => {
  return res.status(200).send("OK");
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err);
  res.status(500).send("Internal Server Error");
});

app.listen(port, () => {
  console.log(`Server started at http://0.0.0.0:${port}`);
});
