package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/kiali/kiali/business"
	"github.com/kiali/kiali/prometheus"
)

// WorkloadList is the API handler to fetch all the workloads to be displayed, related to a single namespace
func WorkloadList(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Get business layer
	business, err := getBusiness(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Workloads initialization error: "+err.Error())
		return
	}
	namespace := params["namespace"]

	// Fetch and build workloads
	workloadList, err := business.Workload.GetWorkloadList(namespace)
	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, workloadList)
}

// WorkloadDetails is the API handler to fetch all details to be displayed, related to a single workload
func WorkloadDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	query := r.URL.Query()

	// Get business layer
	business, err := getBusiness(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Workloads initialization error: "+err.Error())
		return
	}
	namespace := params["namespace"]
	workload := params["workload"]
	workloadType := query.Get("type")

	// Fetch and build workload
	workloadDetails, err := business.Workload.GetWorkload(namespace, workload, workloadType, true)
	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, workloadDetails)
}

// WorkloadUpdate is the API to perform a patch on a Workload configuration
func WorkloadUpdate(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	query := r.URL.Query()

	// Get business layer
	business, err := getBusiness(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Workloads initialization error: "+err.Error())
		return
	}

	namespace := params["namespace"]
	workload := params["workload"]
	workloadType := query.Get("type")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Update request with bad update patch: "+err.Error())
	}
	jsonPatch := string(body)
	workloadDetails, err := business.Workload.UpdateWorkload(namespace, workload, workloadType, true, jsonPatch)

	if err != nil {
		handleErrorResponse(w, err)
		return
	}
	audit(r, "UPDATE on Namespace: "+namespace+" Workload name: "+workload+" Type: "+workloadType+" Patch: "+jsonPatch)
	RespondWithJSON(w, http.StatusOK, workloadDetails)
}

// WorkloadMetrics is the API handler to fetch metrics to be displayed, related to a single workload
func WorkloadMetrics(w http.ResponseWriter, r *http.Request) {
	getWorkloadMetrics(w, r, defaultPromClientSupplier)
}

// getWorkloadMetrics (mock-friendly version)
func getWorkloadMetrics(w http.ResponseWriter, r *http.Request, promSupplier promClientSupplier) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	workload := vars["workload"]

	prom, namespaceInfo := initClientsForMetrics(w, r, promSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, Workload: workload}
	err := extractIstioMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics := prom.GetMetrics(&params)
	RespondWithJSON(w, http.StatusOK, metrics)
}

// WorkloadDashboard is the API handler to fetch Istio dashboard, related to a single workload
func WorkloadDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	workload := vars["workload"]

	prom, namespaceInfo := initClientsForMetrics(w, r, defaultPromClientSupplier, namespace)
	if prom == nil {
		// any returned value nil means error & response already written
		return
	}

	params := prometheus.IstioMetricsQuery{Namespace: namespace, Workload: workload}
	err := extractIstioMetricsQueryParams(r, &params, namespaceInfo)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	svc := business.NewDashboardsService(prom)
	dashboard, err := svc.GetIstioDashboard(params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJSON(w, http.StatusOK, dashboard)
}

// PodDetails is the API handler to fetch all details to be displayed, related to a single pod
func PodDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Get business layer
	business, err := getBusiness(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Pods initialization error: "+err.Error())
		return
	}
	namespace := vars["namespace"]
	pod := vars["pod"]

	// Fetch and build pod
	podDetails, err := business.Workload.GetPod(namespace, pod)
	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, podDetails)
}

// PodLogs is the API handler to fetch logs for a single pod container
func PodLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryParams := r.URL.Query()

	// Get business layer
	business, err := getBusiness(r)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Pod Logs initialization error: "+err.Error())
		return
	}
	namespace := vars["namespace"]
	pod := vars["pod"]

	// Get log options
	opts, err := business.Workload.BuildLogOptionsCriteria(
		queryParams.Get("container"),
		queryParams.Get("duration"),
		queryParams.Get("sinceTime"),
		queryParams.Get("tailLines"))

	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	// Fetch pod logs
	podLogs, err := business.Workload.GetPodLogs(namespace, pod, opts)
	if err != nil {
		handleErrorResponse(w, err)
		return
	}

	RespondWithJSON(w, http.StatusOK, podLogs)
}
