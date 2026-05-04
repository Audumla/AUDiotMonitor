FROM python:3.12-slim

WORKDIR /app

RUN pip install --no-cache-dir flask prometheus-client requests

COPY app.py .

EXPOSE 9300

CMD ["python", "app.py"]
