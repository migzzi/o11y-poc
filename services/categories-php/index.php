<?php

use OpenTelemetry\API\Globals;
use Monolog\Logger;
use OpenTelemetry\Contrib\Logs\Monolog\Handler;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Factory\AppFactory;
use Psr\Http\Server\RequestHandlerInterface as RequestHandler;
use Psr\Log\LogLevel;
use Slim\Routing\RouteCollectorProxy;

require __DIR__ . '/vendor/autoload.php';



$app = AppFactory::create();
// $tracer = Globals::tracerProvider()->getTracer('categories-php');

$loggerProvider = Globals::loggerProvider();
$otlpHandler = new Handler(
    $loggerProvider,
    LogLevel::INFO
);
$stoutHandler = new Monolog\Handler\StreamHandler('php://stdout', LogLevel::INFO);
$monolog = new Logger('categories-service-logger', [$otlpHandler, $stoutHandler]);

$categories = [
    [
        "id" => 1,
        "name" => "Category 1",
        "description" => "Description 1"
    ],
    [
        "id" => 2,
        "name" => "Category 2",
        "description" => "Description 2"
    ],
    [
        "id" => 3,
        "name" => "Category 3",
        "description" => "Description 3"
    ]
];



function getCategoriesNextId(array $categories): int
{
    return max(array_map(function ($category) {
        return $category['id'];
    }, $categories)) + 1;
}

function randomLatencyMW(int $min, int $max)
{
    return function (Request $request, RequestHandler $handler) use ($min, $max) {
        usleep(random_int($min, $max) * 1000);
        return $handler->handle($request);
    };
}


function failWithProbabilityMW(int $chance)
{
    return function (Request $request, RequestHandler $handler) use ($chance) {
        if (random_int(0, 100) < $chance) {
            throw new Exception("Failed with probability of $chance%");
        }
        return $handler->handle($request);
    };
}



$app->get('/might-fail', function (Request $request, Response $response) {
    $response->getBody()->write("Hello");
    return $response;
})->add(failWithProbabilityMW(30));



$app->group('/categories', function (RouteCollectorProxy $group) use ($categories, $monolog) {
    $group->get('', function (Request $request, Response $response) use ($categories) {
        $response->getBody()->write(json_encode($categories));
        return $response->withHeader('Content-Type', 'application/json');
    })->add(failWithProbabilityMW(5));


    $group->get('/{id}', function (Request $request, Response $response, $args) use ($categories) {
        $id = $args['id'];
        $category = array_filter($categories, function ($category) use ($id) {
            return $category['id'] == $id;
        });
        if (empty($category)) {
            $response->getBody()->write(json_encode(["error" => "Category not found"]));
            return $response->withStatus(404);
        }
        $response->getBody()->write(json_encode(array_values($category)[0]));
        return $response->withHeader('Content-Type', 'application/json');
    })->add(failWithProbabilityMW(2));


    $group->post('', function (Request $request, Response $response) use ($categories, $monolog) {
        $data = json_decode($request->getBody(), true);
        $category = [
            "id" => getCategoriesNextId($categories),
            "name" => $data['name'],
            "description" => $data['description']

        ];
        $categories[] = $category;

        $monolog->info("Category created", ["category" => $category]);

        $response->getBody()->write(json_encode($categories));
        return $response->withHeader('Content-Type', 'application/json');
    })->add(failWithProbabilityMW(5));
})->add(randomLatencyMW(0, 300));



$errorMW = $app->addErrorMiddleware(true, true, true);
$errorMW->setDefaultErrorHandler(function (Request $request, Throwable $exception, bool $displayErrorDetails, bool $logErrors, bool $logErrorDetails) use ($monolog) {
    $response = new \Slim\Psr7\Response();
    $monolog->error("Error: " . $exception->getMessage(), ["exception" => $exception]);
    $response->getBody()->write(json_encode(["error" => $exception->getMessage()]));
    return $response->withHeader('Content-Type', 'application/json');
});

//echo env var
$app->get('/env', function (Request $request, Response $response) {
    $response->getBody()->write(json_encode(getenv()));
    return $response->withHeader('Content-Type', 'application/json');
});

$app->run();
